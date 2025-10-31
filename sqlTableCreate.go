package sqlfilestore

import "github.com/dracory/sb"

func (st *Store) sqlTableCreate() string {
	sql := sb.NewBuilder(sb.DatabaseDriverName(st.db)).
		Table(st.tableName).
		Column(sb.Column{
			Name:       COLUMN_ID,
			Type:       sb.COLUMN_TYPE_STRING,
			Length:     40,
			PrimaryKey: true,
		}).
		Column(sb.Column{
			Name:   COLUMN_PARENT_ID,
			Type:   sb.COLUMN_TYPE_STRING,
			Length: 40,
		}).
		Column(sb.Column{
			Name:   COLUMN_TYPE,
			Type:   sb.COLUMN_TYPE_STRING,
			Length: 10,
		}).
		Column(sb.Column{
			Name:   COLUMN_NAME,
			Type:   sb.COLUMN_TYPE_STRING,
			Length: 100,
		}).
		Column(sb.Column{
			Name: COLUMN_CONTENTS,
			Type: sb.COLUMN_TYPE_LONGTEXT,
		}).
		Column(sb.Column{
			Name: COLUMN_SIZE,
			Type: sb.COLUMN_TYPE_INTEGER,
		}).
		Column(sb.Column{
			Name:   COLUMN_EXTENSION,
			Type:   sb.COLUMN_TYPE_STRING,
			Length: 12,
		}).
		Column(sb.Column{
			Name:   COLUMN_PATH,
			Type:   sb.COLUMN_TYPE_STRING,
			Length: 2048,
			// Unique: true,
		}).
		Column(sb.Column{
			Name: COLUMN_CREATED_AT,
			Type: sb.COLUMN_TYPE_DATETIME,
		}).
		Column(sb.Column{
			Name: COLUMN_UPDATED_AT,
			Type: sb.COLUMN_TYPE_DATETIME,
		}).
		Column(sb.Column{
			Name: COLUMN_DELETED_AT,
			Type: sb.COLUMN_TYPE_DATETIME,
		}).
		CreateIfNotExists()

	return sql
}
