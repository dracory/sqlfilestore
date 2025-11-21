package sqlfilestore

import (
	"context"
	"database/sql"
	"errors"

	"github.com/dracory/sb"
)

// NewStoreOptions define the options for creating a new block store
type NewStoreOptions struct {
	TableName          string
	DB                 *sql.DB
	DbDriverName       string
	AutomigrateEnabled bool
	DebugEnabled       bool
}

// NewStore creates a new block store
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
