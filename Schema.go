package builder

import "fmt"

type Schema struct {
	init bool
}



func (c *Schema)ColumnBigInt(length int)*Schema{
	fmt.Println(c.GetInit())
	data := c.Transform()
	fmt.Println(data)
	return data
}
func (c *Schema)ColumnBigPk(length int)*Schema{
	fmt.Println(c.GetInit())
	data := c.Transform()
	fmt.Println(data)
	return data
}
func (c *Schema)ColumnBinary(length int)*Schema{
	data := c.Transform()

	return data
}
func (c *Schema)ColumnBoolean()*Schema{
	data := c.Transform()

	return data
}
func (c *Schema)ColumnDate()*Schema{
	data := c.Transform()

	return data
}
func (c *Schema)ColumnDateTime()*Schema{
	data := c.Transform()

	return data
}
func (c *Schema)ColumnDecimal(precision float32)*Schema{
	data := c.Transform()

	return data
}
func (c *Schema)ColumnDouble(precision float32)*Schema{
	data := c.Transform()

	return data
}
func (c *Schema)ColumnFloat(precision float32)*Schema{
	data := c.Transform()

	return data
}
func (c *Schema)ColumnInt(length int)*Schema{
	data := c.Transform()

	return data
}
func (c *Schema)ColumnMoney(precision float32)*Schema{
	data := c.Transform()

	return data
}
func (c *Schema)ColumnPrimaryKey(length int)*Schema{
	data := c.Transform()

	return data
}
func (c *Schema)ColumnSmallint(length int)*Schema{
	data := c.Transform()

	return data
}
func (c *Schema)ColumnString(length int)*Schema{
	data := c.Transform()

	return data
}
func (c *Schema)ColumnText()*Schema{
	data := c.Transform()

	return data
}
func (c *Schema)ColumnTime(precision float32)*Schema{
	data := c.Transform()

	return data
}
func (c *Schema)ColumnTimestamp(precision float32)*Schema{
	data := c.Transform()

	return data
}
func (c *Schema)ColumnTinyint(length int)*Schema{
	data := c.Transform()

	return data
}

func (c *Schema)returnColumn()string{

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
