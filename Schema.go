package builder

type Schema struct {
	init *Schema
}




func (c *Schema)ColumnBigInt(length int)*Schema{
	data := c.transform()

	return data
}
func (c *Schema)ColumnBigPk(length int)*Schema{
	data := c.transform()

	return data
}
func (c *Schema)ColumnBinary(length int)*Schema{
	data := c.transform()

	return data
}
func (c *Schema)ColumnBoolean()*Schema{
	data := c.transform()

	return data
}
func (c *Schema)ColumnDate()*Schema{
	data := c.transform()

	return data
}
func (c *Schema)ColumnDateTime()*Schema{
	data := c.transform()

	return data
}
func (c *Schema)ColumnDecimal(precision float32)*Schema{
	data := c.transform()

	return data
}
func (c *Schema)ColumnDouble(precision float32)*Schema{
	data := c.transform()

	return data
}
func (c *Schema)ColumnFloat(precision float32)*Schema{
	data := c.transform()

	return data
}
func (c *Schema)ColumnInt(length int)*Schema{
	data := c.transform()

	return data
}
func (c *Schema)ColumnMoney(precision float32)*Schema{
	data := c.transform()

	return data
}
func (c *Schema)ColumnPrimaryKey(length int)*Schema{
	data := c.transform()

	return data
}
func (c *Schema)ColumnSmallint(length int)*Schema{
	data := c.transform()

	return data
}
func (c *Schema)ColumnString(length int)*Schema{
	data := c.transform()

	return data
}
func (c *Schema)ColumnText()*Schema{
	data := c.transform()

	return data
}
func (c *Schema)ColumnTime(precision float32)*Schema{
	data := c.transform()

	return data
}
func (c *Schema)ColumnTimestamp(precision float32)*Schema{
	data := c.transform()

	return data
}
func (c *Schema)ColumnTinyint(length int)*Schema{
	data := c.transform()

	return data
}

func (c *Schema)ReturnColumn(length int)string{

	return ""
}

func (c *Schema)GetInit() bool{
	return true
}

func (c *Schema)transform() *Schema{
	if c.GetInit()==false {
		c.init = &Schema{}
	}
	return c.init
}
