package builder

type AddColumn interface {
	returnColumn() string
}

//type SchemaInterface interface {
//	ColumnBigInt(int)*SchemaInterface
//	ColumnBigPk(int)*SchemaInterface
//	ColumnBinary(int)*SchemaInterface
//	ColumnBoolean()*SchemaInterface
//	ColumnDate()*SchemaInterface
//	ColumnDateTime()*SchemaInterface
//	ColumnDecimal(float32)*SchemaInterface
//	ColumnDouble(float32)*SchemaInterface
//	ColumnFloat(float32)*SchemaInterface
//	ColumnInt(int)*SchemaInterface
//	ColumnMoney(float32)*SchemaInterface
//	ColumnPrimaryKey(int)*SchemaInterface
//	ColumnSmallint(int)*SchemaInterface
//	ColumnString(int)*SchemaInterface
//	ColumnText()*SchemaInterface
//	ColumnTime(float32)*SchemaInterface
//	ColumnTimestamp(float32)*SchemaInterface
//	ColumnTinyint(int)*SchemaInterface
//	GetInit()bool
//}
