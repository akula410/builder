package builder

type Schema struct {
	init bool
	column map[string]interface{}
}



func (c *Schema)ColumnBigInt(length int)*Schema{
	data := c.Transform()
	data.column[SchemaTypeBigint]=length
	return data
}
func (c *Schema)ColumnBigPk(length int)*Schema{
	data := c.Transform()
	data.column[SchemaTypeBigPk]=length
	return data
}
func (c *Schema)ColumnBinary(length int)*Schema{
	data := c.Transform()
	data.column[SchemaTypeBinary]=length
	return data
}
func (c *Schema)ColumnBoolean()*Schema{
	data := c.Transform()
	data.column[SchemaTypeBoolean]=true
	return data
}
func (c *Schema)ColumnDate()*Schema{
	data := c.Transform()
	data.column[SchemaTypeDate]=true
	return data
}
func (c *Schema)ColumnDateTime()*Schema{
	data := c.Transform()
	data.column[SchemaTypeDateTime]=true
	return data
}
func (c *Schema)ColumnDecimal(precision float32)*Schema{
	data := c.Transform()
	data.column[SchemaTypeDecimal]=precision
	return data
}
func (c *Schema)ColumnDouble(precision float32)*Schema{
	data := c.Transform()
	data.column[SchemaTypeDouble]=precision
	return data
}
func (c *Schema)ColumnFloat(precision float32)*Schema{
	data := c.Transform()
	data.column[SchemaTypeFloat]=precision
	return data
}
func (c *Schema)ColumnInt(length int)*Schema{
	data := c.Transform()
	data.column[SchemaTypeInteger]=length
	return data
}
func (c *Schema)ColumnMoney(precision float32)*Schema{
	data := c.Transform()
	data.column[SchemaTypeMoney]=precision
	return data
}
func (c *Schema)ColumnPrimaryKey(length int)*Schema{
	data := c.Transform()
	data.column[SchemaTypePk]=length
	return data
}
func (c *Schema)ColumnSmallint(length int)*Schema{
	data := c.Transform()
	data.column[SchemaTypeSmallint]=length
	return data
}
func (c *Schema)ColumnString(length int)*Schema{
	data := c.Transform()
	data.column[SchemaTypeString] = length
	return data
}
func (c *Schema)ColumnText()*Schema{
	data := c.Transform()
	data.column[SchemaTypeText] = true
	return data
}
func (c *Schema)ColumnTime(precision float32)*Schema{
	data := c.Transform()
	data.column[SchemaTypeTime]=precision
	return data
}
func (c *Schema)ColumnTimestamp(precision float32)*Schema{
	data := c.Transform()
	data.column[SchemaTypeTimeStamp] = precision
	return data
}
func (c *Schema)ColumnTinyint(length int)*Schema{
	data := c.Transform()
	data.column[SchemaTypeTinyint] = length
	return data
}

func (c *Schema)Name(name string)*Schema{
	data := c.Transform()
	data.column[SchemaName] = name
	return data
}

func (c *Schema)NotNull(flag bool)*Schema{
	data := c.Transform()
	data.column[SchemaNotNull] = flag
	return data
}

func (c *Schema)DefaultValue(value interface{})*Schema{
	data := c.Transform()
	data.column[SchemaDefaultValue] = value
	return data
}

func (c *Schema)Comment(value string)*Schema{
	data := c.Transform()
	data.column[SchemaComment] = value
	return data
}

func (c *Schema)Unique()*Schema{
	data := c.Transform()
	data.column[SchemaUnique] = true
	return data
}

func (c *Schema)Index(field string)*Schema{
	data := c.Transform()
	data.column[SchemaIndex] = field
	return data
}

func (c *Schema)IndexName(name string)*Schema{
	data := c.Transform()
	data.column[SchemaIndexName] = name
	return data
}




func (c *Schema)returnColumn()string{

	return ""
}

func (c *Schema)returnIndex()string{

	return ""
}

func (c *Schema)GetInit() bool{
	return c.init
}

func (c *Schema)Transform() *Schema{
	if c.GetInit()==false {
		data := &Schema{}
		data.init = true
		return data
	}else{
		return c
	}
}
