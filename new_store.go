package sqlfilestore

import (
	"context"
	"database/sql"
	"errors"

	"github.com/dracory/neat"
)

// NewStoreOptions defines the configuration options for creating a new Store.
type NewStoreOptions struct {
	TableName          string
	DB                 *sql.DB
	AutomigrateEnabled bool
	DebugEnabled       bool
}

// NewStore creates a new file store instance with the provided options.
func NewStore(opts NewStoreOptions) (*Store, error) {
	if opts.TableName == "" {
		return nil, errors.New("file store: FileTableName is required")
	}

	if opts.DB == nil {
		return nil, errors.New("shop store: DB is required")
	}

	neatDB, err := neat.NewFromSQLDB(opts.DB)
	if err != nil {
		return nil, err
	}

	store := &Store{
		tableName:          opts.TableName,
		automigrateEnabled: opts.AutomigrateEnabled,
		db:                 neatDB,
		debugEnabled:       opts.DebugEnabled,
	}

	if store.automigrateEnabled {
		err := store.MigrateUp(context.Background())

		if err != nil {
			return nil, err
		}
	}

	return store, nil
}
