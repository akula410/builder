package builder

const Select = 1
const From = 2
const Field = 3
const AND = 4
const OR = 5
const IN = 6
const NotIn = 7
const InValue = 8
const InMdf = 9
const NotInMdf = 10
const InValueMdf = 11
const JoinInner = 12
const JoinLeft = 13
const JoinLink = 14
const BktStart = 15
const BktEnd = 16
const Update = 17
const UpdateSet = 18
const Incr = 19
const Decr = 20
const Columns = 21
const ColumnPrimaryKey = 22
const ColumnsIndex = 23
const Limit = 24
const Offset = 25
const OrderBy = 26
const GroupBy = 27



const SchemaTypePk        = "pk"
const SchemaTypeBigPk     = "bigpk"
const SchemaTypeString    = "string"
const SchemaTypeText      = "text"
const SchemaTypeTinyint   = "tinyint"
const SchemaTypeSmallint  = "smallint"
const SchemaTypeInteger   = "integer"
const SchemaTypeBigint    = "bigint"
const SchemaTypeFloat     = "float"
const SchemaTypeDouble    = "double"
const SchemaTypeDecimal   = "decimal"
const SchemaTypeDateTime  = "datetime"
const SchemaTypeTimeStamp = "timestamp"
const SchemaTypeTime      = "time"
const SchemaTypeDate      = "date"
const SchemaTypeBinary    = "binary"
const SchemaTypeBoolean   = "boolean"

const SchemaName         = "name"
const SchemaNotNull      = "notnull"
const SchemaDefaultValue = "default_value"
const SchemaComment = "comment"
const SchemaUnsigned = "unsigned"

const SchemaUnique = "unique"
const SchemaIndex = "index"
const SchemaIndexName = "index_name"


const ConfigCurrentTimestamp = "CURRENT_TIMESTAMP"
const ConfigUpdateCurrentTimestamp = "CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"


const CHARSET_UTF8 = "utf8"

const COLLATE_UTF8_GENERAL_CI = "utf8_general_ci"

const ASC = "ASC"
const DESC = "DESC"

