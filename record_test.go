package sqlfilestore

import (
	"testing"

	"github.com/dracory/sb"
)

func TestNewRecordSetsDefaultValues(t *testing.T) {
	record := NewRecord()

	if record == nil {
		t.Fatal("expected record instance")
	}

	if record.ID() == "" {
		t.Fatal("expected generated id")
	}

	if record.CreatedAt() == "" {
		t.Fatal("expected created_at to be set")
	}

	if record.UpdatedAt() == "" {
		t.Fatal("expected updated_at to be set")
	}

	if record.DeletedAt() != sb.NULL_DATETIME {
		t.Fatalf("expected deleted_at to default to %q", sb.NULL_DATETIME)
	}

	if record.IsDirectory() {
		t.Fatal("new record should not be directory by default")
	}

	if record.IsFile() {
		t.Fatal("new record should not be file by default")
	}
}

func TestNewFileSetsFileType(t *testing.T) {
	record := NewFile()

	if record.Type() != TYPE_FILE {
		t.Fatalf("expected type %q, got %q", TYPE_FILE, record.Type())
	}

	if !record.IsFile() {
		t.Fatal("expected IsFile to be true")
	}

	if record.IsDirectory() {
		t.Fatal("expected IsDirectory to be false for file")
	}
}

func TestNewDirectorySetsExpectedDefaults(t *testing.T) {
	record := NewDirectory()

	if record.Type() != TYPE_DIRECTORY {
		t.Fatalf("expected type %q, got %q", TYPE_DIRECTORY, record.Type())
	}

	if record.Size() != "0" {
		t.Fatalf("expected size to be 0, got %q", record.Size())
	}

	if record.Contents() != "" {
		t.Fatalf("expected contents to be empty, got %q", record.Contents())
	}

	if record.Extension() != "" {
		t.Fatalf("expected extension to be empty, got %q", record.Extension())
	}

	if !record.IsDirectory() {
		t.Fatal("expected IsDirectory to be true")
	}

	if record.IsFile() {
		t.Fatal("expected IsFile to be false for directory")
	}
}

func TestRecordSettersAndGetters(t *testing.T) {
	record := NewRecord()

	if record.SetName("file.txt") != record {
		t.Fatal("SetName should return the same record for chaining")
	}

	record.SetParentID("parent").
		SetPath("  dir/file.txt  ").
		SetContents("hello").
		SetSize("5").
		SetExtension("txt").
		SetType(TYPE_FILE)

	if record.Name() != "file.txt" {
		t.Fatalf("unexpected name: %q", record.Name())
	}

	if record.ParentID() != "parent" {
		t.Fatalf("unexpected parent id: %q", record.ParentID())
	}

	if record.Path() != "/dir/file.txt" {
		t.Fatalf("unexpected path: %q", record.Path())
	}

	if record.Contents() != "hello" {
		t.Fatalf("unexpected contents: %q", record.Contents())
	}

	if record.Size() != "5" {
		t.Fatalf("unexpected size: %q", record.Size())
	}

	if record.Extension() != "txt" {
		t.Fatalf("unexpected extension: %q", record.Extension())
	}

	if !record.IsFile() {
		t.Fatal("record should be marked as file")
	}

	if record.IsDirectory() {
		t.Fatal("record should not be marked as directory")
	}
}

func TestNewRecordFromExistingDataHydratesAllFields(t *testing.T) {
	data := map[string]string{
		"id":         "custom-id",
		"path":       "/dir/file.txt",
		"type":       TYPE_FILE,
		"parent_id":  "parent",
		"name":       "file.txt",
		"contents":   "payload",
		"size":       "42",
		"extension":  "txt",
		"created_at": "2024-01-01 00:00:00",
		"updated_at": "2024-01-02 00:00:00",
		"deleted_at": "2024-01-03 00:00:00",
	}

	record := NewRecordFromExistingData(data)

	if record.ID() != data["id"] {
		t.Fatalf("expected id %q, got %q", data["id"], record.ID())
	}

	if record.Path() != data["path"] {
		t.Fatalf("expected path %q, got %q", data["path"], record.Path())
	}

	if record.Type() != data["type"] {
		t.Fatalf("expected type %q, got %q", data["type"], record.Type())
	}

	if record.ParentID() != data["parent_id"] {
		t.Fatalf("expected parent id %q, got %q", data["parent_id"], record.ParentID())
	}

	if record.Name() != data["name"] {
		t.Fatalf("expected name %q, got %q", data["name"], record.Name())
	}

	if record.Contents() != data["contents"] {
		t.Fatalf("expected contents %q, got %q", data["contents"], record.Contents())
	}

	if record.Size() != data["size"] {
		t.Fatalf("expected size %q, got %q", data["size"], record.Size())
	}

	if record.Extension() != data["extension"] {
		t.Fatalf("expected extension %q, got %q", data["extension"], record.Extension())
	}

	if record.CreatedAt() != data["created_at"] {
		t.Fatalf("expected created_at %q, got %q", data["created_at"], record.CreatedAt())
	}

	if record.UpdatedAt() != data["updated_at"] {
		t.Fatalf("expected updated_at %q, got %q", data["updated_at"], record.UpdatedAt())
	}

	if record.DeletedAt() != data["deleted_at"] {
		t.Fatalf("expected deleted_at %q, got %q", data["deleted_at"], record.DeletedAt())
	}
}
