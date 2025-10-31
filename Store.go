package sqlfilestore

import (
	"database/sql"
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/dracory/sb"
	"github.com/dromara/carbon/v2"
	"github.com/samber/lo"
)

// var _ StoreInterface = (*Store)(nil) // verify it extends the interface

type Store struct {
	tableName          string
	db                 *sql.DB
	dbDriverName       string
	automigrateEnabled bool
	debugEnabled       bool
}

// AutoMigrate auto migrate
func (store *Store) AutoMigrate() error {
	sql := store.sqlTableCreate()

	if sql == "" {
		return errors.New("record table create sql is empty")
	}

	_, err := store.db.Exec(sql)

	if err != nil {
		return err
	}

	recordCount, err := store.RecordCount(RecordQueryOptions{
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

	err = store.RecordCreate(rootDir)

	if err != nil {
		return err
	}

	return nil
}

// EnableDebug - enables the debug option
func (st *Store) EnableDebug(debug bool) {
	st.debugEnabled = debug
}

func (store *Store) RecordRecalculatePath(record *Record, parentRecord *Record) error {
	if record == nil {
		return errors.New("record is nil")
	}

	if parentRecord == nil {
		parentRecord, err := store.RecordFindByID(record.ParentID(), RecordQueryOptions{Columns: []string{"id", "path"}})

		if err != nil {
			return err
		}

		if parentRecord == nil {
			return errors.New("parent record not found")
		}
	}

	record.SetPath(parentRecord.Path() + PATH_SEPARATOR + record.Name())

	err := store.RecordUpdate(record)

	if err != nil {
		return err
	}

	children, err := store.RecordList(RecordQueryOptions{
		ParentID: record.ID(),
		Columns:  []string{"id", "path"},
	})

	if err != nil {
		return err
	}

	for _, child := range children {
		err = store.RecordRecalculatePath(&child, record)

		if err != nil {
			return err
		}
	}

	return nil
}

func (store *Store) RecordCreate(record *Record) error {
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

	_, err := store.db.Exec(sqlStr, params...)

	if err != nil {
		return err
	}

	record.MarkAsNotDirty()

	return nil
}

func (st *Store) RecordCount(options RecordQueryOptions) (int64, error) {
	options.CountOnly = true
	q := st.recordQuery(options)

	sqlStr, _, errSql := q.Limit(1).Select(goqu.COUNT(goqu.Star()).As("count")).ToSQL()

	if errSql != nil {
		return -1, nil
	}

	if st.debugEnabled {
		log.Println(sqlStr)
	}

	db := sb.NewDatabase(st.db, st.dbDriverName)
	mapped, err := db.SelectToMapString(sqlStr)
	if err != nil {
		return -1, err
	}

	if len(mapped) < 1 {
		return -1, nil
	}

	countStr := mapped[0]["count"]

	i, err := strconv.ParseInt(countStr, 10, 64)

	if err != nil {
		return -1, err

	}

	return i, nil
}

func (store *Store) RecordDelete(record *Record) error {
	if record == nil {
		return errors.New("record is nil")
	}

	return store.RecordDeleteByID(record.ID())
}

func (store *Store) RecordDeleteByID(id string) error {
	if id == "" {
		return errors.New("record id is empty")
	}

	subsCount, err := store.RecordCount(RecordQueryOptions{
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

	_, err = store.db.Exec(sqlStr, params...)

	return err
}

func (store *Store) RecordFindByPath(path string, options RecordQueryOptions) (*Record, error) {
	if path == "" {
		return nil, errors.New("record path is empty")
	}

	path = store.fixPath(path)

	options.Path = path
	options.Limit = 1

	list, err := store.RecordList(options)

	if err != nil {
		return nil, err
	}

	if len(list) > 0 {
		return &list[0], nil
	}

	return nil, nil
}

func (store *Store) RecordFindByID(id string, options RecordQueryOptions) (*Record, error) {
	if id == "" {
		return nil, errors.New("record id is empty")
	}

	options.ID = id
	options.Limit = 1

	list, err := store.RecordList(options)

	if err != nil {
		return nil, err
	}

	if len(list) > 0 {
		return &list[0], nil
	}

	return nil, nil
}

func (store *Store) RecordList(options RecordQueryOptions) ([]Record, error) {
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

	db := sb.NewDatabase(store.db, store.dbDriverName)
	modelMaps, err := db.SelectToMapString(sqlStr)
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

func (store *Store) RecordSoftDelete(record *Record) error {
	if record == nil {
		return errors.New("record is nil")
	}

	record.SetDeletedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))

	return store.RecordUpdate(record)
}

func (store *Store) RecordSoftDeleteByID(id string) error {
	record, err := store.RecordFindByID(id, RecordQueryOptions{Columns: []string{"id", "deleted_at"}})

	if err != nil {
		return err
	}

	return store.RecordSoftDelete(record)
}

func (store *Store) RecordUpdate(record *Record) error {
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

	_, err := store.db.Exec(sqlStr, params...)

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
	if strings.HasPrefix(path, PATH_SEPARATOR) {
		return path
	}

	return PATH_SEPARATOR + path
}

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
