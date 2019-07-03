package builder

import (
	"database/sql"
	"fmt"
	"strings"
)

var Conn func()*sql.DB
var ConnClose func()

type Query struct {
	alias string
	dataForBuilding []map[int]interface{}
}

func (c *Query) CreateDatabase(name string, charset string, collation string, notExists bool) sql.Result{
	var textSql string
	if notExists == true {
		textSql = fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET %s COLLATE %s", name, charset, collation)
	}else{
		textSql = fmt.Sprintf("CREATE DATABASE %s CHARACTER SET %s COLLATE %s", name, charset, collation)
	}
	return c.Exec(textSql)
}

func (c *Query) Exec(textSql string) sql.Result{
	res, err := Conn().Exec(textSql)
	if err != nil {
		panic(err.Error())
	}
	return res
}


func (c *Query) Alias(alias string) *Query{
	c.alias = alias
	return c
}

func (c *Query) Select(fields string) *Query{
	c.dataForBuilding = append(c.dataForBuilding, c.trf(SELECT, fields))
	return c
}

func (c *Query) From(table string) *Query{
	c.dataForBuilding = append(c.dataForBuilding, c.trf(FROM, table))
	return c
}

func (c *Query) Where(field string, value interface{}) *Query{
	c.dataForBuilding = append(c.dataForBuilding, c.trf(FIELD, field))
	c.dataForBuilding = append(c.dataForBuilding, c.trf(AND, value))
	return c
}

func (c *Query) WhereOr(field string, value interface{}) *Query{
	c.dataForBuilding = append(c.dataForBuilding, c.trf(FIELD, field))
	c.dataForBuilding = append(c.dataForBuilding, c.trf(OR, value))
	return c
}

func (c *Query) WhereIn(field string, value []interface{}) *Query{
	c.dataForBuilding = append(c.dataForBuilding, c.trf(IN, field))
	c.dataForBuilding = append(c.dataForBuilding, c.trf(IN_VALUE, c.generateInValue(value)))
	return c
}

func (c *Query) WhereNotIn(field string, value []interface{}) *Query{
	c.dataForBuilding = append(c.dataForBuilding, c.trf(NOT_IN, field))
	c.dataForBuilding = append(c.dataForBuilding, c.trf(IN_VALUE, c.generateInValue(value)))
	return c
}

func (c *Query) WhereInModify(field string, value string) *Query{
	c.dataForBuilding = append(c.dataForBuilding, c.trf(IN_MDF, field))
	c.dataForBuilding = append(c.dataForBuilding, c.trf(IN_VALUE_MDF, value))
	return c
}

func (c *Query) WhereNotInModify(field string, value string) *Query{
	c.dataForBuilding = append(c.dataForBuilding, c.trf(NOT_IN_MDF, field))
	c.dataForBuilding = append(c.dataForBuilding, c.trf(IN_VALUE_MDF, value))
	return c
}

func (c *Query) JoinInner(table string, link string) *Query{
	c.dataForBuilding = append(c.dataForBuilding, c.trf(JOIN_INNER, table))
	c.dataForBuilding = append(c.dataForBuilding, c.trf(JOIN_LINK, link))
	return c
}

func (c *Query) JoinLeft(table string, link string) *Query{
	c.dataForBuilding = append(c.dataForBuilding, c.trf(JOIN_LEFT, table))
	c.dataForBuilding = append(c.dataForBuilding, c.trf(JOIN_LINK, link))
	return c
}

func (c *Query) BktStart() *Query{
	c.dataForBuilding = append(c.dataForBuilding, c.trf(BKT_START, true))
	return c
}

func (c *Query) BktEnd() *Query{
	c.dataForBuilding = append(c.dataForBuilding, c.trf(BKT_END, true))
	return c
}

func (c *Query) Update(table string, data ...map[string]interface{}) *Query {
	c.dataForBuilding = append(c.dataForBuilding, c.trf(UPDATE, c.mysqlRealEscapeString(table)))
	var UpdateSet []string
	for _, r := range data {
		for field, result := range r{
			UpdateSet = append(UpdateSet, fmt.Sprintf("%v='%v'", c.mysqlRealEscapeString(field), c.mysqlRealEscapeString(result)))
		}
	}
	c.dataForBuilding = append(c.dataForBuilding, c.trf(UPDATE_SET, strings.Join(UpdateSet, ", ")))

	return c
}

func (c *Query) Incr(data ...interface{}) *Query{
	if len(data) == 1 {
		c.dataForBuilding = append(c.dataForBuilding, c.trf(INCR, c.mysqlRealEscapeString(data[0])))
	}else{
		for i, value := range data {
			if i==0 {
				c.dataForBuilding = append(c.dataForBuilding, c.trf(UPDATE, c.mysqlRealEscapeString(value)))
			} else {
				c.dataForBuilding = append(c.dataForBuilding, c.trf(INCR, c.mysqlRealEscapeString(value)))
			}
		}
	}

	return c
}

func (c *Query) Decr(data ...interface{}) *Query{
	if len(data) == 1 {
		c.dataForBuilding = append(c.dataForBuilding, c.trf(DECR, c.mysqlRealEscapeString(data[0])))
	}else{
		for i, value := range data {
			if i==0 {
				c.dataForBuilding = append(c.dataForBuilding, c.trf(UPDATE, c.mysqlRealEscapeString(value)))
			} else {
				c.dataForBuilding = append(c.dataForBuilding, c.trf(DECR, c.mysqlRealEscapeString(value)))
			}
		}
	}

	return c
}



func (c *Query) Build() string{
	sqlRequest := ""
	var SelectSlice       []string
	var FromSlice         []string
	var JoinSlice         []string
	var WhereSlice        []string
	var UpdateSlice         string
	var UpdateParamsSlice []string


	for _, box := range c.dataForBuilding{
		for key, value := range box{
			switch key{
			case SELECT:
				SelectSlice = append(SelectSlice, fmt.Sprintf("%v", c.mysqlRealEscapeString(value)))

			case FROM:
				FromSlice = append(FromSlice, fmt.Sprintf("%v", value))

			case FIELD:
				var field string
				if value == nil{
				field = fmt.Sprintf("%v", c.mysqlRealEscapeString(value))
			} else{
				field = fmt.Sprintf("%v", c.mysqlRealEscapeString(value))
			}
				WhereSlice = append(WhereSlice, field)

			case AND:
				if value==nil {
					WhereSlice = append(WhereSlice, "IS NULL ")
				}else{
					WhereSlice = append(WhereSlice, fmt.Sprintf("='%v'", c.mysqlRealEscapeString(value)))
				}

				WhereSlice = append(WhereSlice, "AND")

			case OR:
				if value==nil {
					WhereSlice = append(WhereSlice, "IS NULL ")
				}else{
					WhereSlice = append(WhereSlice, fmt.Sprintf("='%v'", c.mysqlRealEscapeString(value)))
				}
				WhereSlice = append(WhereSlice, "OR")

			case IN:
				WhereSlice = append(WhereSlice, fmt.Sprintf("%v IN ", c.mysqlRealEscapeString(value)))

			case NOT_IN:
				WhereSlice = append(WhereSlice, fmt.Sprintf("%v NOT IN ", c.mysqlRealEscapeString(value)))

			case IN_VALUE:
				WhereSlice = append(WhereSlice, fmt.Sprintf("%v", value))
				WhereSlice = append(WhereSlice, "AND")

			case IN_MDF:
				WhereSlice = append(WhereSlice, fmt.Sprintf("%v IN ", value))

			case NOT_IN_MDF:
				WhereSlice = append(WhereSlice, fmt.Sprintf("%v NOT IN ", value))

			case IN_VALUE_MDF:
				WhereSlice = append(WhereSlice, fmt.Sprintf("(%v)", value))
				WhereSlice = append(WhereSlice, "AND")

			case JOIN_INNER:
				JoinSlice = append(JoinSlice, fmt.Sprintf("INNER JOIN %v", value))

			case JOIN_LEFT:
				JoinSlice = append(JoinSlice, fmt.Sprintf("LEFT JOIN %v", value))

			case JOIN_LINK:
				JoinSlice = append(JoinSlice, fmt.Sprintf(" ON %v", value))

			case BKT_START:
				WhereSlice = append(WhereSlice, "(")
			case BKT_END:
				WhereSlice = append(WhereSlice, ")")
			case UPDATE:
				UpdateSlice = fmt.Sprintf("UPDATE %v", value)
			case UPDATE_SET:
				UpdateParamsSlice = append(UpdateParamsSlice, fmt.Sprintf(" %s ", value))
			case INCR:
				UpdateParamsSlice = append(UpdateParamsSlice, fmt.Sprintf(" %s = %s+1 ", value, value))
			case DECR:
				UpdateParamsSlice = append(UpdateParamsSlice, fmt.Sprintf(" %s = %s-1 ", value, value))
			}
		}
	}

	if len(SelectSlice)>0 {
		sqlRequest += fmt.Sprintf("SELECT %s ", strings.Join(SelectSlice, ", "))
	}

	if len(UpdateSlice)>0 {
		sqlRequest += fmt.Sprintf("%s SET %s", UpdateSlice, strings.Join(UpdateParamsSlice, ", "))
	}

	if len(FromSlice)>0 {
		sqlRequest += fmt.Sprintf("FROM %s ", strings.Join(FromSlice, ", "))
	}

	if len(JoinSlice)>0 {
		sqlRequest += strings.Join(JoinSlice, " ")
	}

	if len(WhereSlice)>0 {

		sqlRequest += fmt.Sprintf(" WHERE %s", strings.Join(c.mdfWhere(WhereSlice), " "))
	}

	if len(c.alias)>0 {
		sqlRequest = fmt.Sprintf("(%s) AS %s", sqlRequest, c.alias)
	}

	return sqlRequest
}

func (c *Query) Apply()int64{
	var aff int64
	if textSql := c.Build(); textSql!=""{
		res, err := Conn().Exec(textSql)
		if err != nil {
			panic(err.Error())
		}

		aff, err = res.RowsAffected()
		if err != nil {
			panic(err.Error())
		}

	}

	return aff
}

func (c *Query) Rows() []map[string]interface{}{
	var result []map[string]interface{}
	if textSql := c.Build(); textSql!=""{
		rows, err := Conn().Query(textSql)

		if err != nil {
			panic(err.Error())
		}
		columns, err := rows.Columns()

		if err != nil {
			panic(err.Error())
		}

		count := len(columns)
		values := make([]interface{}, count)
		valuePtrs := make([]interface{}, count)

		for rows.Next() {
			for i := range columns {
				valuePtrs[i] = &values[i]
			}

			err := rows.Scan(valuePtrs...)
			if err != nil {
				panic(err.Error())
			}

			fields := make(map[string]interface{})
			for i, col := range columns {

				var v interface{}

				val := values[i]

				b, ok := val.([]byte)

				if ok {
					v = string(b)
				} else {
					v = val
				}
				fields[col] = v
			}
			result = append(result, fields)
		}

		err = rows.Close()
		if err != nil {
			panic(err)
		}
	}


	return result
}

func (c *Query) Row() map[string]interface{} {
	var result map[string]interface{}

	if textSql := c.Build(); textSql!=""{

		rows, err := Conn().Query(textSql)

		if err != nil {
			panic(err.Error())
		}
		columns, err := rows.Columns()

		if err != nil {
			panic(err.Error())
		}

		count := len(columns)
		values := make([]interface{}, count)
		valuePtrs := make([]interface{}, count)

		for rows.Next() {
			for i := range columns {
				valuePtrs[i] = &values[i]
			}

			err := rows.Scan(valuePtrs...)
			if err != nil {
				panic(err.Error())
			}

			fields := make(map[string]interface{})
			for i, col := range columns {

				var v interface{}

				val := values[i]

				b, ok := val.([]byte)

				if ok {
					v = string(b)
				} else {
					v = val
				}
				fields[col] = v
			}
			result = fields
			break
		}
		err = rows.Close()
		if err != nil {
			panic(err)
		}
	}


	return result
}

func (c *Query) Insert(table string, data []map[string]interface{}) int64{
	/** Названия полей таблицы */
	var dataField  []string

	/** Структура значений таблицы */
	var dataValue =  make([]string, 0, len(data))

	/** Значения таблицы */
	var dataBox    []interface{}

	/** Для задания порядка чтения карты */
	var fields []string

	var textSql string

	var res sql.Result

	var aff int64


	/** Наполнение структуры данными */
	for i, box := range data {
		dataBoxValue := make([]string, 0, len(box))
		if i==0 {
			dataField = make([]string, 0, len(box))
			fields = make([]string, 0, len(box))
			for field, v := range box {
				dataBox = append(dataBox, v)
				dataBoxValue = append(dataBoxValue, "?")
				dataField = append(dataField, field)
				fields = append(fields, field)
			}
		}else{
			for _, field := range fields{
				dataBox = append(dataBox, box[field])
				dataBoxValue = append(dataBoxValue, "?")
			}
		}

		dataValue = append(dataValue, fmt.Sprintf("(%s)", strings.Join(dataBoxValue, ", ")))
	}

	if len(dataField)>0 {
		/** Формирование запроса */
		textSql = fmt.Sprintf("INSERT INTO %s(%s) VALUES %s", table, strings.Join(dataField, ", "), strings.Join(dataValue, ", "))

		ins, err := Conn().Prepare(textSql)
		if err != nil {
			panic(err.Error())
		}

		res, err = ins.Exec(dataBox...)

		if err != nil {
			panic(err.Error())
		}

		aff, err = res.RowsAffected()

		if err != nil {
			panic(err.Error())
		}

		err = ins.Close()
		if err!= nil {
			panic(err)
		}
	}

	return aff
}

func (c *Query) Delete() {

}

func (c *Query) trf(key int, value interface{}) map[int]interface{}{
	trf := make(map[int]interface{})
	trf[key] = value
	return trf
}

func (c *Query) generateInValue(value []interface{}) string {
	inValue := make([]string, 0, len(value))

	for _, v := range value{
		inValue = append(inValue, fmt.Sprintf("'%v'", c.mysqlRealEscapeString(v)))
	}
	return fmt.Sprintf(" (%s) ", strings.Join(inValue, ", "))
}

func (c *Query) mdfWhere(where []string) []string{

	var deleteKey int
	var whereResult []string
	for i, v := range where{
		if v=="AND" || v=="OR"{
			deleteKey = i
		}
	}

	if deleteKey>0 {
		for i, v := range where{
			if deleteKey != i{
				whereResult = append(whereResult, v)
			}
		}
	}

	return whereResult
}

func (c *Query) mysqlRealEscapeString(value interface{}) string {
	strValue := fmt.Sprintf("%v", value)
	replace := map[string]string{"\\":"\\\\", "'":`\'`, "\\0":"\\\\0", "\n":"\\n", "\r":"\\r", `"`:`\"`, "\x1a":"\\Z"}

	for b, a := range replace {
		strValue = strings.Replace(strValue, b, a, -1)
	}

	return strValue
}
