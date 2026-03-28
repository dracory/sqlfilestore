package sqlfilestore

import (
	"context"
	"database/sql"
	"errors"

	"github.com/dracory/sb"
)

// NewStoreOptions defines the configuration options for creating a new Store.
// All fields are optional except TableName and DB which are required.
type NewStoreOptions struct {
	TableName          string
	DB                 *sql.DB
	DbDriverName       string
	AutomigrateEnabled bool
	DebugEnabled       bool
}

// NewStore creates a new file store instance with the provided options.
//
// Required options:
//   - TableName: the database table name for storing records
//   - DB: a *sql.DB database connection
//
// Optional options:
//   - DbDriverName: auto-detected if not provided
//   - AutomigrateEnabled: if true, creates table and root directory automatically
//   - DebugEnabled: if true, logs all SQL statements
//
// Returns an error if required options are missing or if auto-migration fails.
func NewStore(opts NewStoreOptions) (*Store, error) {
	if opts.TableName == "" {
		return nil, errors.New("file store: FileTableName is required")
	}

	if opts.DB == nil {
		return nil, errors.New("shop store: DB is required")
	}

	if opts.DbDriverName == "" {
		opts.DbDriverName = sb.DatabaseDriverName(opts.DB)
	}

	store := &Store{
		tableName:          opts.TableName,
		automigrateEnabled: opts.AutomigrateEnabled,
		db:                 opts.DB,
		dbDriverName:       opts.DbDriverName,
		debugEnabled:       opts.DebugEnabled,
	}

	if store.automigrateEnabled {
		err := store.AutoMigrate(context.Background())

		if err != nil {
			return nil, err
		}
	}

	return store, nil
}
