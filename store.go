package sqlfilestore

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/dracory/neat"
	contractsorm "github.com/dracory/neat/contracts/database/orm"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/dromara/carbon/v2"
)

// Store provides a hierarchical file-system-like storage in a SQL database.
type Store struct {
	tableName          string
	db                 *neat.Database
	automigrateEnabled bool
	debugEnabled       bool
}

// MigrateUp creates the database table if it doesn't exist and ensures
// a root directory record is present.
func (store *Store) MigrateUp(ctx context.Context, tx ...*sql.Tx) error {
	if store.db.Schema().HasTable(store.tableName) {
		return nil
	}

	err := store.db.Schema().Create(store.tableName, func(table contractsschema.Blueprint) {
		table.String(COLUMN_ID, 40)
		table.Primary(COLUMN_ID)
		table.String(COLUMN_PARENT_ID, 40)
		table.String(COLUMN_TYPE, 10)
		table.String(COLUMN_NAME, 100)
		table.Text(COLUMN_CONTENTS)
		table.String(COLUMN_SIZE, 20)
		table.String(COLUMN_EXTENSION, 12)
		table.String(COLUMN_PATH, 2048)
		table.DateTime(COLUMN_CREATED_AT)
		table.DateTime(COLUMN_UPDATED_AT)
		table.DateTime(COLUMN_SOFT_DELETED_AT)
	})

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

// MigrateDown drops the sqlfilestore table
func (store *Store) MigrateDown(ctx context.Context, tx ...*sql.Tx) error {
	if !store.db.Schema().HasTable(store.tableName) {
		return nil
	}

	err := store.db.Schema().Drop(store.tableName)
	if err != nil {
		return err
	}

	return nil
}

// EnableDebug enables or disables SQL debug logging.
func (st *Store) EnableDebug(debug bool) {
	st.debugEnabled = debug
	if debug {
		st.db.EnableDebug()
	} else {
		st.db.DisableDebug()
	}
}

// RecordRecalculatePath updates the path of a record and all its children
// after a rename or move operation.
func (store *Store) RecordRecalculatePath(ctx context.Context, record *Record, parentRecord *Record) error {
	if record == nil {
		return errors.New("record is nil")
	}

	if parentRecord == nil {
		parentRecord, err := store.RecordFindByID(ctx, record.ParentID(), RecordQueryOptions{})

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
func (store *Store) RecordCreate(ctx context.Context, record *Record) error {
	record.SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))
	record.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))

	row := map[string]any{
		COLUMN_ID:              record.ID(),
		COLUMN_PARENT_ID:       record.ParentID(),
		COLUMN_NAME:            record.Name(),
		COLUMN_PATH:            record.Path(),
		COLUMN_TYPE:            record.Type(),
		COLUMN_SIZE:            record.Size(),
		COLUMN_EXTENSION:       record.Extension(),
		COLUMN_CONTENTS:        record.Contents(),
		COLUMN_CREATED_AT:      record.GetCreatedAtCarbon().StdTime(),
		COLUMN_UPDATED_AT:      record.GetUpdatedAtCarbon().StdTime(),
		COLUMN_SOFT_DELETED_AT: record.GetDeletedAtCarbon().StdTime(),
	}

	return store.db.Query().Table(store.tableName).Create(row)
}

// RecordCount returns the count of records matching the provided query options.
func (st *Store) RecordCount(ctx context.Context, options RecordQueryOptions) (int64, error) {
	q := st.buildQuery(options)

	var count int64
	err := q.Table(st.tableName).Count(&count)
	return count, err
}

// RecordDelete permanently removes a record from the database.
func (store *Store) RecordDelete(ctx context.Context, record *Record) error {
	if record == nil {
		return errors.New("record is nil")
	}

	return store.RecordDeleteByID(ctx, record.ID())
}

// RecordDeleteByID permanently deletes a record by its ID.
func (store *Store) RecordDeleteByID(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("record id is empty")
	}

	subsCount, err := store.RecordCount(ctx, RecordQueryOptions{
		ParentID:        id,
		WithSoftDeleted: true,
	})

	if err != nil {
		return err
	}

	if subsCount > 0 {
		return errors.New("directory is not empty")
	}

	_, err = store.db.Query().
		Table(store.tableName).
		Where(COLUMN_ID+" = ?", id).
		Delete()

	return err
}

// RecordFindByPath finds a single record by its path.
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
func (store *Store) RecordList(ctx context.Context, options RecordQueryOptions) ([]Record, error) {
	q := store.buildQuery(options)

	type recordRow struct {
		ID            string    `db:"id"`
		ParentID      string    `db:"parent_id"`
		Name          string    `db:"name"`
		Path          string    `db:"path"`
		Type          string    `db:"type"`
		Size          string    `db:"size"`
		Extension     string    `db:"extension"`
		Contents      string    `db:"contents"`
		CreatedAt     time.Time `db:"created_at"`
		UpdatedAt     time.Time `db:"updated_at"`
		SoftDeletedAt time.Time `db:"soft_deleted_at"`
	}

	var rows []recordRow
	if err := q.Table(store.tableName).Get(&rows); err != nil {
		return []Record{}, err
	}

	list := make([]Record, 0, len(rows))
	for _, r := range rows {
		rec := &Record{}
		rec.SetID(r.ID)
		rec.SetParentID(r.ParentID)
		rec.SetName(r.Name)
		rec.SetPath(r.Path)
		rec.SetType(r.Type)
		rec.SetSize(r.Size)
		rec.SetExtension(r.Extension)
		rec.SetContents(r.Contents)
		rec.CreatedAtField.CreatedAt = r.CreatedAt
		rec.UpdatedAtField.UpdatedAt = r.UpdatedAt
		rec.SoftDeletesMaxDate.SoftDeletedAt = r.SoftDeletedAt
		list = append(list, *rec)
	}

	return list, nil
}

// RecordSoftDelete marks a record as deleted by setting the deleted_at timestamp.
func (store *Store) RecordSoftDelete(ctx context.Context, record *Record) error {
	if record == nil {
		return errors.New("record is nil")
	}

	record.SetDeletedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))

	return store.RecordUpdate(ctx, record)
}

// RecordSoftDeleteByID finds a record by ID and marks it as soft deleted.
func (store *Store) RecordSoftDeleteByID(ctx context.Context, id string) error {
	record, err := store.RecordFindByID(ctx, id, RecordQueryOptions{})

	if err != nil {
		return err
	}

	return store.RecordSoftDelete(ctx, record)
}

// RecordUpdate updates an existing record in the database.
func (store *Store) RecordUpdate(ctx context.Context, record *Record) error {
	if record == nil {
		return errors.New("record is nil")
	}

	record.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString())

	row := map[string]any{
		COLUMN_PARENT_ID:       record.ParentID(),
		COLUMN_NAME:            record.Name(),
		COLUMN_PATH:            record.Path(),
		COLUMN_TYPE:            record.Type(),
		COLUMN_SIZE:            record.Size(),
		COLUMN_EXTENSION:       record.Extension(),
		COLUMN_CONTENTS:        record.Contents(),
		COLUMN_UPDATED_AT:      record.GetUpdatedAtCarbon().StdTime(),
		COLUMN_SOFT_DELETED_AT: record.GetDeletedAtCarbon().StdTime(),
	}

	_, err := store.db.Query().
		Table(store.tableName).
		Where(COLUMN_ID+" = ?", record.ID()).
		Update(row)

	return err
}

func (store *Store) buildQuery(options RecordQueryOptions) contractsorm.Query {
	q := store.db.Query()

	if options.ID != "" {
		q = q.Where(COLUMN_ID+" = ?", options.ID)
	}

	if len(options.IDIn) > 0 {
		args := make([]any, len(options.IDIn))
		for i, id := range options.IDIn {
			args[i] = id
		}
		q = q.WhereIn(COLUMN_ID, args)
	}

	if options.ParentID != "" {
		q = q.Where(COLUMN_PARENT_ID+" = ?", options.ParentID)
	}

	if options.CreatedAtGreaterThan != "" {
		q = q.Where(COLUMN_CREATED_AT+" > ?", options.CreatedAtGreaterThan)
	}

	if options.CreatedAtLessThan != "" {
		q = q.Where(COLUMN_CREATED_AT+" < ?", options.CreatedAtLessThan)
	}

	if options.UpdatedAtGreaterThan != "" {
		q = q.Where(COLUMN_UPDATED_AT+" > ?", options.UpdatedAtGreaterThan)
	}

	if options.UpdatedAtLessThan != "" {
		q = q.Where(COLUMN_UPDATED_AT+" < ?", options.UpdatedAtLessThan)
	}

	if options.Type != "" {
		q = q.Where(COLUMN_TYPE+" = ?", options.Type)
	}

	if options.Path != "" {
		q = q.Where(COLUMN_PATH+" = ?", options.Path)
	}

	if options.PathStartsWith != "" {
		q = q.Where(COLUMN_PATH+" LIKE ?", options.PathStartsWith+"%")
	}

	if options.Limit > 0 {
		q = q.Limit(options.Limit)
	}

	if options.Offset > 0 {
		q = q.Offset(options.Offset)
	}

	sortOrder := "desc"
	if options.SortOrder != "" {
		sortOrder = options.SortOrder
	}

	if options.OrderBy != "" {
		q = q.OrderBy(options.OrderBy, sortOrder)
	}

	if options.WithSoftDeleted {
		q = q.WithSoftDeleted()
	} else {
		q = q.Where(COLUMN_SOFT_DELETED_AT+" = ?", carbon.Parse(MAX_DATETIME, carbon.UTC).StdTime())
	}

	return q
}

func (store *Store) fixPath(path string) string {
	path = strings.ReplaceAll(path, "\\", PATH_SEPARATOR)

	if strings.HasPrefix(path, PATH_SEPARATOR) {
		return path
	}

	return PATH_SEPARATOR + path
}

// RecordQueryOptions defines the available filters and options for record queries.
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
