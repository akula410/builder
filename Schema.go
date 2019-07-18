package builder

import (
	"fmt"
	"strings"
)

type Schema struct {
	init bool
	column map[string]interface{}
}


func (c *Schema)ColumnPrimaryKey(length int)*Schema{
	data := c.Transform()
	data.column[SchemaTypePk]=length
	return data
}
func (c *Schema)ColumnBigPrimaryKey(length int)*Schema{
	data := c.Transform()
	data.column[SchemaTypeBigPk]=length
	return data
}
func (c *Schema)ColumnBigInt(length int)*Schema{
	data := c.Transform()
	data.column[SchemaTypeBigint]=length
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



func (c *Schema)Unsigned()*Schema{
	data := c.Transform()
	data.column[SchemaUnsigned] = true
	return data
}

func (c *Schema)Unique()*Schema{
	data := c.Transform()
	data.column[SchemaUnique] = true
	return data
}

func (c *Schema)Index()*Schema{
	data := c.Transform()
	data.column[SchemaIndex] = true
	return data
}

func (c *Schema)IndexName(name string)*Schema{
	data := c.Transform()
	data.column[SchemaIndexName] = name
	return data
}




func (c *Schema)returnColumn()(string, string, string){
	sqlText := make([]string, 0)
	var columnName string
	var columnType string
	var columnDefault string
	var columnNull string
	var columnAutoIncrement string
	var columnComment string
	var columnUnsigned bool
	var columnPrimaryKey string
	var columnIndex string

	//Init
	columnNull = "NULL"

	fmt.Println(c.column)
	for key, value := range c.column {
		switch key {
		case SchemaTypePk:
			columnPrimaryKey = fmt.Sprintf("%v", c.column[SchemaName])
			columnIndex = fmt.Sprintf("%v", c.column[SchemaName])
			columnType = fmt.Sprintf("INT(%v)", value)
			columnNull = "NOT NULL"
			columnAutoIncrement = "AUTO_INCREMENT"

		case SchemaTypeBigPk:
			columnPrimaryKey = fmt.Sprintf("%v", c.column[SchemaName])
			columnIndex = fmt.Sprintf("%v", c.column[SchemaName])
			columnType = fmt.Sprintf("BIGINT(%v)", value)
			columnNull = "NOT NULL"
			columnAutoIncrement = "AUTO_INCREMENT"

		case SchemaTypeString:
			columnType = fmt.Sprintf("VARCHAR(%v)", value)

		case SchemaTypeText:
			columnType = "TEXT"

		case SchemaTypeTinyint:
			columnType = fmt.Sprintf("TINYINT(%v)", value)

		case SchemaTypeSmallint:
			columnType = fmt.Sprintf("SMALLINT(%v)", value)

		case SchemaTypeInteger:
			columnType = fmt.Sprintf("INTEGER(%v)", value)

		case SchemaTypeBigint:
			columnType = fmt.Sprintf("BIGINT(%v)", value)

		case SchemaTypeFloat:
			columnType = fmt.Sprintf("FLOAT(%s)", strings.Replace(fmt.Sprintf("%v", value), ".", ",", 1))

		case SchemaTypeDouble:
			columnType = fmt.Sprintf("DOUBLE(%s)", strings.Replace(fmt.Sprintf("%v", value), ".", ",", 1))

		case SchemaTypeDecimal:
			columnType = fmt.Sprintf("DECIMAL(%s)", strings.Replace(fmt.Sprintf("%v", value), ".", ",", 1))

		case SchemaTypeDateTime:
			columnType = "DATETIME"
		case SchemaTypeTimeStamp:
			columnType = "TIMESTAMP"
			columnType = fmt.Sprintf("TIMESTAMP %v", value)

		case SchemaTypeTime:
			columnType = "TIME"

		case SchemaTypeDate:
			columnType = "DATE"

		case SchemaTypeBinary:
			columnType = fmt.Sprintf("BINARY(%v)", value)

		case SchemaTypeBoolean:
			columnType = "BOOLEAN"

		case SchemaName:
			columnName = fmt.Sprintf("%v", value)

		case SchemaNotNull:
			if value == true {
				columnNull = "NOT NULL"
			}else{
				columnNull = "NULL"
			}

		case SchemaDefaultValue:
			columnDefault = fmt.Sprintf("DEFAULT '%v'", value)

		case SchemaComment:
			columnComment = fmt.Sprintf("COMMENT '%v'", value)

		case SchemaUnsigned:
			columnUnsigned = true

		case SchemaIndex:
			columnIndex = fmt.Sprintf("%v", c.column[SchemaName])
		}

	}

	if len(columnName)>0 {
		sqlText = append(sqlText, columnName)
	}

	if len(columnType)>0 {
		sqlText = append(sqlText, columnType)
	}

	if len(columnDefault)>0 {
		sqlText = append(sqlText, columnDefault)
	}

	if columnUnsigned || len(columnPrimaryKey)>0 {
		sqlText = append(sqlText, "UNSIGNED")
	}

	if len(columnNull)>0{
		sqlText = append(sqlText, columnNull)
	}else{
		sqlText = append(sqlText, "NULL")
	}

	if len(columnAutoIncrement)>0 {
		sqlText = append(sqlText, columnAutoIncrement)
	}

	if len(columnComment)>0 {
		sqlText = append(sqlText, columnComment)
	}

	if _, ok := c.column[SchemaIndexName]; ok {
		columnIndex = fmt.Sprintf("%v", fmt.Sprintf("%v", c.column[SchemaIndexName]))
	}

	return strings.Join(sqlText, " "), columnPrimaryKey, columnIndex
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
		data.column = make(map[string]interface{})
		return data
	}else{
		return c
	}
}
