# SQL File Store

<a href="https://gitpod.io/#https://github.com/gouniverse/sqlfilestore" style="float:right:"><img src="https://gitpod.io/button/open-in-gitpod.svg" alt="Open in Gitpod" loading="lazy"></a>

![tests](https://github.com/gouniverse/sqlfilestore/workflows/tests/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/gouniverse/sqlfilestore)](https://goreportcard.com/report/github.com/gouniverse/sqlfilestore)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/gouniverse/sqlfilestore)](https://pkg.go.dev/github.com/gouniverse/sqlfilestore)

SQL File Store persists a hierarchical file-system-like structure in a relational database. It handles record creation, querying, soft deletion, and path recalculation for nested directories and files, while keeping a root directory available out of the box.

## Features

- Automatic schema creation (automigration) with a configurable table name.
- Strongly typed `Record` helper with builders for files and directories.
- CRUD helpers: create, update, hard delete, soft delete, and list with filtering options.
- Recursive path recalculation to keep child paths in sync with renamed parents.
- Works with standard `database/sql` connections and supports multiple SQL drivers through query builders.

## Installation

```bash
go get github.com/dracory/sqlfilestore
```

## Quick Start

```go
package main

import (
    "database/sql"

    "github.com/dracory/sqlfilestore"
    _ "modernc.org/sqlite" // driver
)

func main() {
    db, _ := sql.Open("sqlite", ":memory:?parseTime=true")

    store, err := sqlfilestore.NewStore(sqlfilestore.NewStoreOptions{
        DB:                 db,
        TableName:          "file_records",
        AutomigrateEnabled: true,
    })
    if err != nil {
        panic(err)
    }

    file := sqlfilestore.NewFile().
        SetParentID(sqlfilestore.ROOT_ID).
        SetName("example.txt").
        SetPath("/example.txt").
        SetExtension("txt").
        SetContents("Hello, world!")

    if err := store.RecordCreate(file); err != nil {
        panic(err)
    }
}
```

The example mirrors the test setup used for the in-memory SQLite driver.

## Working with Records

All records share a common `Record` model. File and directory helpers pre-configure type-specific fields:

- `NewDirectory()` sets type to `directory`, zeroes size, and clears file-only fields.
- `NewFile()` sets type to `file` and leaves contents configurable.
- `NewRecord()` creates a bare record with generated ID and timestamps.

Each record exposes setters/getters for metadata including name, path, parent ID, file size, extension, and timestamps.

### Creating Directories

```go
dir := sqlfilestore.NewDirectory().
    SetParentID(sqlfilestore.ROOT_ID).
    SetName("docs").
    SetPath("/docs")

if err := store.RecordCreate(dir); err != nil {
    // handle error
}
```

### Updating Paths

Rename operations require updating the stored path. Use `RecordRecalculatePath` to refresh a record and its descendants after changing the name.

```go
dir.SetName("manuals")
if err := store.RecordUpdate(dir); err == nil {
    _ = store.RecordRecalculatePath(dir, nil)
}
```

## Querying Data

`RecordQueryOptions` lets you filter by ID, parent, type, path, timestamps, and sorting options.

```go
records, err := store.RecordList(sqlfilestore.RecordQueryOptions{
    ParentID: sqlfilestore.ROOT_ID,
    Type:     sqlfilestore.TYPE_FILE,
    OrderBy:  "created_at",
    SortOrder: "asc",
})
```

Use `RecordFindByID` and `RecordFindByPath` for single-record lookups.

## Soft Deletes vs. Hard Deletes

- `RecordSoftDeleteByID` keeps the record while setting `deleted_at` for reversible removals.
- `RecordDeleteByID` permanently removes a record, ensuring directories are empty beforehand.

Soft-deleted records are excluded by default; enable `WithSoftDeleted` in query options to include them.

## Debugging

Enable SQL logging with `store.EnableDebug(true)` to print generated statements before execution.

## Running Tests

The repository includes comprehensive tests that demonstrate typical interactions with the store.

```bash
go test ./...
```

## License

GPL-3.0. See [LICENSE](LICENSE).
