package sqlfilestore

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/dracory/database"
	"github.com/dracory/sb"
	"github.com/dromara/carbon/v2"
	"github.com/samber/lo"
)

// var _ StoreInterface = (*Store)(nil) // verify it extends the interface

// Store provides a hierarchical file-system-like storage in a SQL database.
// It handles record creation, querying, soft deletion, and path recalculation
// for nested directories and files, while keeping a root directory available
// out of the box.
type Store struct {
	tableName          string
	db                 *sql.DB
	dbDriverName       string
	automigrateEnabled bool
	debugEnabled       bool
}

// AutoMigrate creates the database table if it doesn't exist and ensures
// a root directory record is present. It runs automatically if AutomigrateEnabled
// is set to true in NewStoreOptions.
func (store *Store) AutoMigrate(ctx context.Context) error {
	sql, err := store.sqlTableCreate()

	if err != nil {
		return err
	}

	if sql == "" {
		return errors.New("record table create sql is empty")
	}

	_, err = store.db.Exec(sql)

	if err != nil {
		return err
	}

	recordCount, err := store.RecordCount(ctx, RecordQueryOptions{
		Path: ROOT_PATH,
	})

	if err != nil {
		return err
	}

	if recordCount > 0 {
		return nil
	}

	rootDir := NewDirectory().
		SetID(ROOT_ID).
		SetPath(ROOT_PATH).
		SetName("root").
		SetParentID("-1")

	err = store.RecordCreate(ctx, rootDir)

	if err != nil {
		return err
	}

	return nil
}

// EnableDebug enables or disables SQL debug logging.
// When enabled, generated SQL statements are printed to stdout.
func (st *Store) EnableDebug(debug bool) {
	st.debugEnabled = debug
}

// RecordRecalculatePath updates the path of a record and all its children
// after a rename or move operation. If parentRecord is nil, it looks up the parent.
// This recursively updates all descendant paths to reflect the new parent path.
func (store *Store) RecordRecalculatePath(ctx context.Context, record *Record, parentRecord *Record) error {
	if record == nil {
		return errors.New("record is nil")
	}

	if parentRecord == nil {
		parentRecord, err := store.RecordFindByID(ctx, record.ParentID(), RecordQueryOptions{Columns: []string{"id", "path"}})

		if err != nil {
			return err
		}

		if parentRecord == nil {
			return errors.New("parent record not found")
		}
	}

	record.SetPath(parentRecord.Path() + PATH_SEPARATOR + record.Name())

	err := store.RecordUpdate(ctx, record)

	if err != nil {
		return err
	}

	children, err := store.RecordList(ctx, RecordQueryOptions{
		ParentID: record.ID(),
		Columns:  []string{"id", "path"},
	})

	if err != nil {
		return err
	}

	for _, child := range children {
		err = store.RecordRecalculatePath(ctx, &child, record)

		if err != nil {
			return err
		}
	}

	return nil
}

// RecordCreate inserts a new record into the database.
// It automatically sets created_at and updated_at timestamps.
func (store *Store) RecordCreate(ctx context.Context, record *Record) error {
	record.SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))
	record.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))

	data := record.Data()

	sqlStr, params, errSql := goqu.Dialect(store.dbDriverName).
		Insert(store.tableName).
		Prepared(true).
		Rows(data).
		ToSQL()

	if errSql != nil {
		return errSql
	}

	if store.debugEnabled {
		log.Println(sqlStr)
	}

	_, err := store.db.ExecContext(ctx, sqlStr, params...)

	if err != nil {
		return err
	}

	record.MarkAsNotDirty()

	return nil
}

// RecordCount returns the count of records matching the provided query options.
//
// It builds a COUNT query based on the filter criteria in options (ID, ParentID, Type,
// Path filters, timestamps, soft delete status, etc.). The CountOnly flag is always
// set internally regardless of the input options.
//
// Returns:
//   - (count, nil) on success, where count is the number of matching records (0 or more)
//   - (0, error) if query building fails, database query fails, or count parsing fails
//
// Note: This function never returns negative counts. All error cases return 0 as the count value.
// The original implementation returned -1 on some errors, but this was changed to 0 for consistency.
func (st *Store) RecordCount(ctx context.Context, options RecordQueryOptions) (int64, error) {
	options.CountOnly = true
	q := st.recordQuery(options)

	sqlStr, sqlParams, errSql := q.Limit(1).Select(goqu.COUNT(goqu.Star()).As("count")).ToSQL()

	if errSql != nil {
		return 0, errSql
	}

	if st.debugEnabled {
		log.Println(sqlStr)
	}

	mapped, err := database.SelectToMapString(database.NewQueryableContext(ctx, st.db), sqlStr, sqlParams...)
	if err != nil {
		return 0, err
	}

	if len(mapped) < 1 {
		return 0, nil
	}

	countStr := mapped[0]["count"]

	i, err := strconv.ParseInt(countStr, 10, 64)

	if err != nil {
		return 0, err
	}

	return i, nil
}

// RecordDelete permanently removes a record from the database.
// It delegates to RecordDeleteByID using the record's ID.
func (store *Store) RecordDelete(ctx context.Context, record *Record) error {
	if record == nil {
		return errors.New("record is nil")
	}

	return store.RecordDeleteByID(ctx, record.ID())
}

// RecordDeleteByID permanently deletes a record by its ID.
// It ensures directories are empty before deletion (no children allowed).
// Returns an error if the directory is not empty or if the ID is invalid.
func (store *Store) RecordDeleteByID(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("record id is empty")
	}

	subsCount, err := store.RecordCount(ctx, RecordQueryOptions{
		ParentID:        id,
		CountOnly:       true,
		WithSoftDeleted: true,
	})

	if err != nil {
		return err
	}

	if subsCount > 0 {
		return errors.New("directory is not empty")
	}

	sqlStr, params, errSql := goqu.Dialect(store.dbDriverName).
		Delete(store.tableName).
		Prepared(true).
		Where(goqu.C("id").Eq(id)).
		ToSQL()

	if errSql != nil {
		return errSql
	}

	if store.debugEnabled {
		log.Println(sqlStr)
	}

	_, err = store.db.ExecContext(ctx, sqlStr, params...)

	return err
}

// RecordFindByPath finds a single record by its path.
// Returns nil if no record is found or if the path is empty.
func (store *Store) RecordFindByPath(ctx context.Context, path string, options RecordQueryOptions) (*Record, error) {
	if path == "" {
		return nil, errors.New("record path is empty")
	}

	path = store.fixPath(path)

	options.Path = path
	options.Limit = 1

	list, err := store.RecordList(ctx, options)

	if err != nil {
		return nil, err
	}

	if len(list) > 0 {
		return &list[0], nil
	}

	return nil, nil
}

// RecordFindByID finds a single record by its ID.
// Returns nil if no record is found or if the ID is empty.
func (store *Store) RecordFindByID(ctx context.Context, id string, options RecordQueryOptions) (*Record, error) {
	if id == "" {
		return nil, errors.New("record id is empty")
	}

	options.ID = id
	options.Limit = 1

	list, err := store.RecordList(ctx, options)

	if err != nil {
		return nil, err
	}

	if len(list) > 0 {
		return &list[0], nil
	}

	return nil, nil
}

// RecordList retrieves a list of records matching the provided query options.
// Supports filtering by ID, ParentID, Type, Path, timestamps, and soft delete status.
// Results can be limited, offset, and ordered using the options struct.
func (store *Store) RecordList(ctx context.Context, options RecordQueryOptions) ([]Record, error) {
	q := store.recordQuery(options)

	if len(options.Columns) > 0 {
		q = q.Select(options.Columns[0])
		if len(options.Columns) > 1 {
			for _, column := range options.Columns[1:] {
				q = q.SelectAppend(goqu.C(column))
			}
		}
	} else {
		q = q.Select(goqu.Star())
	}

	sqlStr, _, errSql := q.ToSQL()

	if errSql != nil {
		return []Record{}, nil
	}

	if store.debugEnabled {
		log.Println(sqlStr)
	}

	modelMaps, err := database.SelectToMapString(database.NewQueryableContext(ctx, store.db), sqlStr)
	if err != nil {
		return []Record{}, err
	}

	list := []Record{}

	lo.ForEach(modelMaps, func(modelMap map[string]string, index int) {
		model := NewRecordFromExistingData(modelMap)
		list = append(list, *model)
	})

	return list, nil
}

// RecordSoftDelete marks a record as deleted by setting the deleted_at timestamp.
// The record remains in the database but is excluded from default queries.
func (store *Store) RecordSoftDelete(ctx context.Context, record *Record) error {
	if record == nil {
		return errors.New("record is nil")
	}

	record.SetDeletedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))

	return store.RecordUpdate(ctx, record)
}

// RecordSoftDeleteByID finds a record by ID and marks it as soft deleted.
// Returns an error if the record is not found.
func (store *Store) RecordSoftDeleteByID(ctx context.Context, id string) error {
	record, err := store.RecordFindByID(ctx, id, RecordQueryOptions{Columns: []string{"id", "deleted_at"}})

	if err != nil {
		return err
	}

	return store.RecordSoftDelete(ctx, record)
}

// RecordUpdate updates an existing record in the database.
// It automatically sets the updated_at timestamp and only modifies changed fields.
// The ID field cannot be updated.
func (store *Store) RecordUpdate(ctx context.Context, record *Record) error {
	if record == nil {
		return errors.New("record is nil")
	}

	record.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString())

	dataChanged := record.DataChanged()

	delete(dataChanged, "id") // ID is not updateable

	if len(dataChanged) < 1 {
		return nil
	}

	sqlStr, params, errSql := goqu.Dialect(store.dbDriverName).
		Update(store.tableName).
		Prepared(true).
		Set(dataChanged).
		Where(goqu.C("id").Eq(record.ID())).
		ToSQL()

	if errSql != nil {
		return errSql
	}

	if store.debugEnabled {
		log.Println(sqlStr)
	}

	_, err := store.db.ExecContext(ctx, sqlStr, params...)

	record.MarkAsNotDirty()

	return err
}

func (store *Store) recordQuery(options RecordQueryOptions) *goqu.SelectDataset {
	q := goqu.Dialect(store.dbDriverName).From(store.tableName)

	if options.ID != "" {
		q = q.Where(goqu.C("id").Eq(options.ID))
	}

	if len(options.IDIn) > 0 {
		q = q.Where(goqu.C("id").In(options.IDIn))
	}

	if options.ParentID != "" {
		q = q.Where(goqu.C("parent_id").Eq(options.ParentID))
	}

	if options.CreatedAtGreaterThan != "" {
		q = q.Where(goqu.C("created_at").Gt(options.CreatedAtGreaterThan))
	}

	if options.CreatedAtLessThan != "" {
		q = q.Where(goqu.C("created_at").Lt(options.CreatedAtLessThan))
	}

	if options.UpdatedAtGreaterThan != "" {
		q = q.Where(goqu.C("updated_at").Gt(options.UpdatedAtGreaterThan))
	}

	if options.UpdatedAtLessThan != "" {
		q = q.Where(goqu.C("updated_at").Lt(options.UpdatedAtLessThan))
	}

	if options.Type != "" {
		q = q.Where(goqu.C("type").Eq(options.Type))
	}

	if options.Path != "" {
		q = q.Where(goqu.C("path").Eq(options.Path))
	}

	if options.PathStartsWith != "" {
		q = q.Where(goqu.C("path").Like(options.PathStartsWith + "%"))
	}

	if !options.CountOnly {
		if options.Limit > 0 {
			q = q.Limit(uint(options.Limit))
		}

		if options.Offset > 0 {
			q = q.Offset(uint(options.Offset))
		}
	}

	sortOrder := "desc"
	if options.SortOrder != "" {
		sortOrder = options.SortOrder
	}

	if options.OrderBy != "" {
		if strings.EqualFold(sortOrder, sb.ASC) {
			q = q.Order(goqu.I(options.OrderBy).Asc())
		} else {
			q = q.Order(goqu.I(options.OrderBy).Desc())
		}
	}

	if !options.WithSoftDeleted {
		q = q.Where(goqu.C("deleted_at").Eq(sb.NULL_DATETIME))
	}

	return q
}

func (store *Store) fixPath(path string) string {
	// Normalize backslashes to forward slashes (Windows compatibility)
	path = strings.ReplaceAll(path, "\\", PATH_SEPARATOR)

	if strings.HasPrefix(path, PATH_SEPARATOR) {
		return path
	}

	return PATH_SEPARATOR + path
}

// RecordQueryOptions defines the available filters and options for record queries.
// Use this struct with RecordList, RecordCount, RecordFindByID, and RecordFindByPath.
type RecordQueryOptions struct {
	ID                   string
	IDIn                 []string
	ParentID             string
	Type                 string
	Path                 string
	PathStartsWith       string
	CreatedAtLessThan    string
	CreatedAtGreaterThan string
	UpdatedAtLessThan    string
	UpdatedAtGreaterThan string
	Columns              []string
	Offset               int
	Limit                int
	SortOrder            string
	OrderBy              string
	CountOnly            bool
	WithSoftDeleted      bool
}
