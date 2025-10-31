package sqlfilestore

import (
	"strings"

	"github.com/dracory/dataobject"
	"github.com/dracory/sb"
	"github.com/dracory/uid"
	"github.com/dromara/carbon/v2"
)

// == CLASS ==================================================================

type Record struct {
	dataobject.DataObject
}

// == CONSTRUCTORS ===========================================================

func NewDirectory() *Record {
	record := NewRecord().
		SetType(TYPE_DIRECTORY). // Directories are of TYPE_DIRECTORY
		SetSize("0").            // Directories have no size when created (they are empty)
		SetContents("").         // Directories have no contents
		SetExtension("")         // Directories have no extension

	return record
}

func NewFile() *Record {
	record := NewRecord().
		SetType(TYPE_FILE) // Files are of TYPE_FILE

	return record
}

func NewRecord() *Record {
	o := (&Record{}).
		SetID(uid.HumanUid()).
		// SetPath("").
		// SetType("").
		// SetParentID("").
		// SetName("").
		// SetContents("").
		// SetName("").
		// SetSize("0").
		SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC)).
		SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC)).
		SetDeletedAt(sb.NULL_DATETIME)
	return o
}

func NewRecordFromExistingData(data map[string]string) *Record {
	o := &Record{}
	o.Hydrate(data)
	return o
}

// == HELPER METHODS =========================================================

func (o *Record) IsDirectory() bool {
	return o.Type() == TYPE_DIRECTORY
}

func (o *Record) IsFile() bool {
	return o.Type() == TYPE_FILE
}

// == SETTERS AND GETTERS =====================================================

func (o *Record) Contents() string {
	return o.Get("contents")
}

func (o *Record) SetContents(fileContents string) *Record {
	o.Set("contents", fileContents)
	return o
}

func (o *Record) CreatedAt() string {
	return o.Get("created_at")
}

func (o *Record) SetCreatedAt(createdAt string) *Record {
	o.Set("created_at", createdAt)
	return o
}

func (o *Record) DeletedAt() string {
	return o.Get("deleted_at")
}

func (o *Record) SetDeletedAt(deletedAt string) *Record {
	o.Set("deleted_at", deletedAt)
	return o
}

func (o *Record) Extension() string {
	return o.Get("extension")
}

func (o *Record) SetExtension(extension string) *Record {
	o.Set("extension", extension)
	return o
}

func (o *Record) ID() string {
	return o.Get("id")
}

func (o *Record) SetID(id string) *Record {
	o.Set("id", id)
	return o
}

func (o *Record) Name() string {
	return o.Get("name")
}

func (o *Record) SetName(name string) *Record {
	o.Set("name", name)
	return o
}

func (o *Record) ParentID() string {
	return o.Get("parent_id")
}

func (o *Record) SetParentID(parentID string) *Record {
	o.Set("parent_id", parentID)
	return o
}

func (o *Record) Path() string {
	return o.Get("path")
}

// SetPath sets the file path. As all paths must start with "/"
// adds a "/" if not present.
// Any trailing spaces is also trimmed
func (o *Record) SetPath(filePath string) *Record {
	filePath = strings.TrimSpace(filePath)
	filePath = "/" + strings.TrimLeft(filePath, "/")
	o.Set("path", filePath)
	return o
}

func (o *Record) Size() string {
	return o.Get("size")
}

func (o *Record) SetSize(fileSize string) *Record {
	o.Set("size", fileSize)
	return o
}

func (o *Record) Type() string {
	return o.Get("type")
}

func (o *Record) SetType(fileType string) *Record {
	o.Set("type", fileType)
	return o
}

func (o *Record) UpdatedAt() string {
	return o.Get("updated_at")
}

func (o *Record) SetUpdatedAt(updatedAt string) *Record {
	o.Set("updated_at", updatedAt)
	return o
}
