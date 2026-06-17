package sqlfilestore

import (
	"strings"

	"github.com/dracory/neat/database/orm"
	"github.com/dracory/neat/database/soft_delete"
	neatuid "github.com/dracory/neat/support/uid"
	"github.com/dromara/carbon/v2"
)

// Record represents a file or directory entry in the hierarchical file store.
type Record struct {
	orm.ShortID

	ParentIDField  string `db:"parent_id"`
	NameField      string `db:"name"`
	PathField      string `db:"path"`
	TypeField      string `db:"type"`
	SizeField      string `db:"size"`
	ExtensionField string `db:"extension"`
	ContentsField  string `db:"contents"`

	CreatedAtField orm.CreatedAt
	UpdatedAtField orm.UpdatedAt
	soft_delete.SoftDeletesMaxDate
}

// NewDirectory creates a new Record preconfigured as a directory.
func NewDirectory() *Record {
	record := NewRecord().
		SetType(TYPE_DIRECTORY).
		SetSize("0").
		SetContents("").
		SetExtension("")

	return record
}

// NewFile creates a new Record preconfigured as a file.
func NewFile() *Record {
	record := NewRecord().
		SetType(TYPE_FILE)

	return record
}

// NewRecord creates a new Record with default values.
func NewRecord() *Record {
	o := &Record{}
	o.SetID(neatuid.GenerateShortID())
	o.SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))
	o.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))
	o.SetDeletedAt(MAX_DATETIME)
	return o
}

// NewRecordFromExistingData creates a Record from a map of existing data.
func NewRecordFromExistingData(data map[string]string) *Record {
	o := &Record{}
	if v, ok := data[COLUMN_ID]; ok {
		o.SetID(v)
	}
	if v, ok := data[COLUMN_PARENT_ID]; ok {
		o.SetParentID(v)
	}
	if v, ok := data[COLUMN_NAME]; ok {
		o.SetName(v)
	}
	if v, ok := data[COLUMN_PATH]; ok {
		o.SetPath(v)
	}
	if v, ok := data[COLUMN_TYPE]; ok {
		o.SetType(v)
	}
	if v, ok := data[COLUMN_SIZE]; ok {
		o.SetSize(v)
	}
	if v, ok := data[COLUMN_EXTENSION]; ok {
		o.SetExtension(v)
	}
	if v, ok := data[COLUMN_CONTENTS]; ok {
		o.SetContents(v)
	}
	if v, ok := data[COLUMN_CREATED_AT]; ok {
		o.SetCreatedAt(v)
	}
	if v, ok := data[COLUMN_UPDATED_AT]; ok {
		o.SetUpdatedAt(v)
	}
	if v, ok := data[COLUMN_SOFT_DELETED_AT]; ok {
		o.SetDeletedAt(v)
	}
	return o
}

// IsDirectory returns true if this record represents a directory.
func (o *Record) IsDirectory() bool {
	return o.Type() == TYPE_DIRECTORY
}

// IsFile returns true if this record represents a file.
func (o *Record) IsFile() bool {
	return o.Type() == TYPE_FILE
}

// IsSoftDeleted returns true if the record is soft deleted.
func (o *Record) IsSoftDeleted() bool {
	return o.SoftDeletesMaxDate.IsSoftDeleted()
}

// GetCreatedAtCarbon returns the created at time as a carbon object.
func (o *Record) GetCreatedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(o.CreatedAtField.CreatedAt)
}

// GetUpdatedAtCarbon returns the updated at time as a carbon object.
func (o *Record) GetUpdatedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(o.UpdatedAtField.UpdatedAt)
}

// GetDeletedAtCarbon returns the soft deleted at time as a carbon object.
func (o *Record) GetDeletedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(o.SoftDeletesMaxDate.SoftDeletedAt)
}

// Contents returns the contents of the record.
func (o *Record) Contents() string {
	return o.ContentsField
}

// SetContents sets the contents of the record.
func (o *Record) SetContents(fileContents string) *Record {
	o.ContentsField = fileContents
	return o
}

// CreatedAt returns the created at time of the record.
func (o *Record) CreatedAt() string {
	if o.CreatedAtField.CreatedAt.IsZero() {
		return ""
	}
	return carbon.CreateFromStdTime(o.CreatedAtField.CreatedAt).ToDateTimeString()
}

// SetCreatedAt sets the created at time of the record.
func (o *Record) SetCreatedAt(createdAt string) *Record {
	if createdAt == "" {
		return o
	}
	o.CreatedAtField.CreatedAt = carbon.Parse(createdAt, carbon.UTC).StdTime()
	return o
}

// DeletedAt returns the soft deleted at time of the record.
func (o *Record) DeletedAt() string {
	if o.SoftDeletesMaxDate.SoftDeletedAt.IsZero() {
		return ""
	}
	return carbon.CreateFromStdTime(o.SoftDeletesMaxDate.SoftDeletedAt).ToDateTimeString()
}

// SetDeletedAt sets the soft deleted at time of the record.
func (o *Record) SetDeletedAt(deletedAt string) *Record {
	if deletedAt == "" {
		return o
	}
	o.SoftDeletesMaxDate.SoftDeletedAt = carbon.Parse(deletedAt, carbon.UTC).StdTime()
	return o
}

// Extension returns the extension of the record.
func (o *Record) Extension() string {
	return o.ExtensionField
}

// SetExtension sets the extension of the record.
func (o *Record) SetExtension(extension string) *Record {
	o.ExtensionField = extension
	return o
}

// ID returns the id of the record.
func (o *Record) ID() string {
	return o.ShortID.ID
}

// SetID sets the id of the record.
func (o *Record) SetID(id string) *Record {
	o.ShortID.ID = id
	return o
}

// Name returns the name of the record.
func (o *Record) Name() string {
	return o.NameField
}

// SetName sets the name of the record.
func (o *Record) SetName(name string) *Record {
	o.NameField = name
	return o
}

// ParentID returns the parent id of the record.
func (o *Record) ParentID() string {
	return o.ParentIDField
}

// SetParentID sets the parent id of the record.
func (o *Record) SetParentID(parentID string) *Record {
	o.ParentIDField = parentID
	return o
}

// Path returns the path of the record.
func (o *Record) Path() string {
	return o.PathField
}

// SetPath sets the file path. As all paths must start with "/"
// adds a "/" if not present.
// Any trailing spaces is also trimmed
func (o *Record) SetPath(filePath string) *Record {
	filePath = strings.TrimSpace(filePath)
	filePath = "/" + strings.TrimLeft(filePath, "/")
	o.PathField = filePath
	return o
}

// Size returns the size of the record.
func (o *Record) Size() string {
	return o.SizeField
}

// SetSize sets the size of the record.
func (o *Record) SetSize(fileSize string) *Record {
	o.SizeField = fileSize
	return o
}

// Type returns the type of the record.
func (o *Record) Type() string {
	return o.TypeField
}

// SetType sets the type of the record.
func (o *Record) SetType(fileType string) *Record {
	o.TypeField = fileType
	return o
}

// UpdatedAt returns the updated at time of the record.
func (o *Record) UpdatedAt() string {
	if o.UpdatedAtField.UpdatedAt.IsZero() {
		return ""
	}
	return carbon.CreateFromStdTime(o.UpdatedAtField.UpdatedAt).ToDateTimeString()
}

// SetUpdatedAt sets the updated at time of the record.
func (o *Record) SetUpdatedAt(updatedAt string) *Record {
	if updatedAt == "" {
		return o
	}
	o.UpdatedAtField.UpdatedAt = carbon.Parse(updatedAt, carbon.UTC).StdTime()
	return o
}
