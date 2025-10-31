package sqlfilestore

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	"github.com/dracory/sb"
	"github.com/spf13/cast"
	_ "modernc.org/sqlite"
)

func initDB(filepath string) *sql.DB {
	os.Remove(filepath) // remove database
	dsn := filepath + "?parseTime=true"
	db, err := sql.Open("sqlite", dsn)

	if err != nil {
		panic(err)
	}

	return db
}

func TestStoreRootCreated(t *testing.T) {
	db := initDB(":memory:")

	store, err := NewStore(NewStoreOptions{
		DB:                 db,
		TableName:          "file_table_create",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if store == nil {
		t.Fatal("unexpected nil store")
	}

	rootDir, err := store.RecordFindByPath(ROOT_PATH, RecordQueryOptions{Columns: []string{"id", "path", "type", "name"}})

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if rootDir == nil {
		t.Fatal("unexpected nil record")
	}

	if rootDir.ID() != ROOT_ID {
		t.Fatal("unexpected root id")
	}

	if rootDir.Path() != ROOT_PATH {
		t.Fatal("unexpected root path")
	}

	if rootDir.Type() != TYPE_DIRECTORY {
		t.Fatal("unexpected root type", rootDir.Type())
	}

	if rootDir.Name() != "root" {
		t.Fatal("unexpected root name")
	}

}

func TestStoreFileCreate(t *testing.T) {
	db := initDB(":memory:")

	store, err := NewStore(NewStoreOptions{
		DB:                 db,
		TableName:          "file_table_create",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if store == nil {
		t.Fatal("unexpected nil store")
	}

	file := NewFile().
		SetParentID(ROOT_ID).
		SetType(TYPE_FILE).
		SetName("test.txt").
		SetPath(ROOT_PATH + "test.txt").
		SetSize(cast.ToString(len([]byte("TEST")))).
		SetExtension("txt").
		SetContents("TEST")

	err = store.RecordCreate(file)

	if err != nil {
		t.Fatal("unexpected error:", err)
	}
}

func TestStoreFileFindByID(t *testing.T) {
	db := initDB(":memory:")

	store, err := NewStore(NewStoreOptions{
		DB:                 db,
		TableName:          "file_table_find_by_id",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	file := NewFile().
		SetParentID(ROOT_ID).
		SetType(TYPE_FILE).
		SetName("test.txt").
		SetPath(ROOT_PATH + "test.txt").
		SetSize(cast.ToString(len([]byte("TEST")))).
		SetExtension("txt").
		SetContents("TEST")

	err = store.RecordCreate(file)
	if err != nil {
		t.Error("unexpected error:", err)
	}

	fileFound, errFind := store.RecordFindByID(file.ID(), RecordQueryOptions{})

	if errFind != nil {
		t.Fatal("unexpected error:", errFind)
	}

	if fileFound == nil {
		t.Fatal("File MUST NOT be nil")
	}

	if fileFound.ID() != file.ID() {
		t.Fatal("IDs do not match")
	}

	if fileFound.Path() != file.Path() {
		t.Fatal("File paths do not match")
	}

	if fileFound.Type() != TYPE_FILE {
		t.Fatal("First type do not match", TYPE_FILE, "found:", fileFound.Type())
	}

	if fileFound.Name() != file.Name() {
		t.Fatal("File names do not match", file.Name(), "found:", fileFound.Name())
	}

	if fileFound.Size() != file.Size() {
		t.Fatal("File sizes do not match", file.Size(), "found:", fileFound.Size())
	}

	if fileFound.Extension() != file.Extension() {
		t.Fatal("File extensions do not match", file.Extension(), "found:", fileFound.Extension())
	}

	if fileFound.IsDirectory() {
		t.Fatal("File MUST NOT be a directory")
	}

	if !fileFound.IsFile() {
		t.Fatal("File MUST be a file")
	}
}

func TestStoreFileSoftDelete(t *testing.T) {
	db := initDB(":memory:")

	store, err := NewStore(NewStoreOptions{
		DB:                 db,
		TableName:          "file_soft_delete",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if store == nil {
		t.Fatal("unexpected nil store")
	}

	file := NewFile().
		SetParentID(ROOT_ID).
		SetType(TYPE_FILE).
		SetName("test.txt").
		SetPath(ROOT_PATH + "test.txt").
		SetSize(cast.ToString(len([]byte("TEST")))).
		SetExtension("txt").
		SetContents("TEST")

	err = store.RecordCreate(file)

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	err = store.RecordSoftDeleteByID(file.ID())

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if file.DeletedAt() != sb.NULL_DATETIME {
		t.Fatal("File MUST NOT be soft deleted")
	}

	fileFound, errFind := store.RecordFindByID(file.ID(), RecordQueryOptions{})

	if errFind != nil {
		t.Fatal("unexpected error:", errFind)
	}

	if fileFound != nil {
		t.Fatal("File MUST be nil")
	}

	fileFindWithDeleted, err := store.RecordList(RecordQueryOptions{
		ID:              file.ID(),
		Limit:           1,
		WithSoftDeleted: true,
	})

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if len(fileFindWithDeleted) == 0 {
		t.Fatal("File MUST be soft deleted")
	}

	if strings.Contains(fileFindWithDeleted[0].DeletedAt(), sb.NULL_DATETIME) {
		t.Fatal("File MUST be soft deleted", file.DeletedAt())
	}
}

func TestStoreFileDelete(t *testing.T) {
	db := initDB(":memory:")

	store, err := NewStore(NewStoreOptions{
		DB:                 db,
		TableName:          "file_soft_delete",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if store == nil {
		t.Fatal("unexpected nil store")
	}

	file := NewFile().
		SetParentID(ROOT_ID).
		SetType(TYPE_FILE).
		SetName("test.txt").
		SetPath(ROOT_PATH + "test.txt").
		SetSize(cast.ToString(len([]byte("TEST")))).
		SetExtension("txt").
		SetContents("TEST")

	err = store.RecordCreate(file)

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	err = store.RecordDeleteByID(file.ID())

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	fileFound, errFind := store.RecordFindByID(file.ID(), RecordQueryOptions{})

	if errFind != nil {
		t.Fatal("unexpected error:", errFind)
	}

	if fileFound != nil {
		t.Fatal("File MUST be nil")
	}

	fileFindWithDeleted, err := store.RecordList(RecordQueryOptions{
		ID:              file.ID(),
		Limit:           1,
		WithSoftDeleted: true,
	})

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if len(fileFindWithDeleted) > 0 {
		t.Fatal("File MUST not be found")
	}
}

func TestStoreFolderDeleteWithSubs(t *testing.T) {
	db := initDB(":memory:")

	store, err := NewStore(NewStoreOptions{
		DB:                 db,
		TableName:          "file_soft_delete_with_subs",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if store == nil {
		t.Fatal("unexpected nil store")
	}

	dir := NewDirectory().
		SetParentID(ROOT_ID).
		SetName("testDir").
		SetPath(ROOT_PATH + "testDir")

	err = store.RecordCreate(dir)

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	file := NewFile().
		SetParentID(dir.ID()).
		SetType(TYPE_FILE).
		SetName("test.txt").
		SetPath(dir.Path() + PATH_SEPARATOR + "test.txt").
		SetSize(cast.ToString(len([]byte("TEST")))).
		SetExtension("txt").
		SetContents("TEST")

	err = store.RecordCreate(file)

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	// Delete folder, must fail as it has subs

	err = store.RecordDeleteByID(dir.ID())

	if err == nil {
		t.Fatal("must return error as directory is not empty")
	}

	if err.Error() != "directory is not empty" {
		t.Fatal("unexpected error:", err)
	}

	// Now delete subs, then delete directory

	err = store.RecordDeleteByID(file.ID())

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	err = store.RecordDeleteByID(dir.ID())

	if err != nil {
		t.Fatal("unexpected error:", err)
	}
}
