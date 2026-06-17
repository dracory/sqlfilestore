package sqlfilestore

// PATH_SEPARATOR is the path delimiter used throughout the file store.
const PATH_SEPARATOR = "/"

// TYPE_FILE is the record type value for files.
const TYPE_FILE = "file"

// TYPE_DIRECTORY is the record type value for directories.
const TYPE_DIRECTORY = "directory"

// ROOT_PATH is the path of the root directory.
const ROOT_PATH = PATH_SEPARATOR

// ROOT_ID is the ID of the root directory record.
const ROOT_ID = "0"

// Database column names used in the file records table.
const (
	COLUMN_ID              = "id"
	COLUMN_PARENT_ID       = "parent_id"
	COLUMN_NAME            = "name"
	COLUMN_PATH            = "path"
	COLUMN_TYPE            = "type"
	COLUMN_SIZE            = "size"
	COLUMN_EXTENSION       = "extension"
	COLUMN_CONTENTS        = "contents"
	COLUMN_CREATED_AT      = "created_at"
	COLUMN_UPDATED_AT      = "updated_at"
	COLUMN_SOFT_DELETED_AT = "soft_deleted_at"
)

// MAX_DATETIME is a far-future datetime used as the default soft-delete sentinel.
const MAX_DATETIME = "9999-12-31 23:59:59"
