package builder

import (
	"database/sql"
	"fmt"
	"strings"
)

var Conn func()*sql.DB
var ConnClose func()

type Query struct {
	Schema

	alias string
	dataForBuilding []map[int]interface{}

	schemaColumns []string
	schemaIndex []string
	schemaPrimaryKey string

	schemaAfterCreateTable func(*Query)
	schemaBeforeCreateTable func(*Query)

	schemaAfterDeleteTable func(*Query)
	schemaBeforeDeleteTable func(*Query)

	schemaTableName string

	tableEngine string

	scriptToConsole bool
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

	c.toSendConsole(textSql)

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
	c.dataForBuilding = append(c.dataForBuilding, c.trf(Select, fields))
	return c
}

func (c *Query) From(table string) *Query{
	c.dataForBuilding = append(c.dataForBuilding, c.trf(From, table))
	return c
}

func (c *Query) Between(field string, value1 interface{}, value2 interface{}){
	c.dataForBuilding = append(c.dataForBuilding, c.trf(Field, field))
	c.dataForBuilding = append(c.dataForBuilding, c.trf(BetweenValue1, value1))
	c.dataForBuilding = append(c.dataForBuilding, c.trf(BetweenValue2, value2))
}

func (c *Query) Where(field string, value interface{}) *Query{
	c.dataForBuilding = append(c.dataForBuilding, c.trf(Field, field))
	c.dataForBuilding = append(c.dataForBuilding, c.trf(AND, value))
	return c
}

func (c *Query) WhereOr(field string, value interface{}) *Query{
	c.dataForBuilding = append(c.dataForBuilding, c.trf(Field, field))
	c.dataForBuilding = append(c.dataForBuilding, c.trf(OR, value))
	return c
}

func (c *Query) WhereIn(field string, value []interface{}) *Query{
	c.dataForBuilding = append(c.dataForBuilding, c.trf(IN, field))
	c.dataForBuilding = append(c.dataForBuilding, c.trf(InValue, c.generateInValue(value)))
	return c
}

func (c *Query) WhereNotIn(field string, value []interface{}) *Query{
	c.dataForBuilding = append(c.dataForBuilding, c.trf(NotIn, field))
	c.dataForBuilding = append(c.dataForBuilding, c.trf(InValue, c.generateInValue(value)))
	return c
}

func (c *Query) WhereInModify(field string, value string) *Query{
	c.dataForBuilding = append(c.dataForBuilding, c.trf(InMdf, field))
	c.dataForBuilding = append(c.dataForBuilding, c.trf(InValueMdf, value))
	return c
}

func (c *Query) WhereNotInModify(field string, value string) *Query{
	c.dataForBuilding = append(c.dataForBuilding, c.trf(NotInMdf, field))
	c.dataForBuilding = append(c.dataForBuilding, c.trf(InValueMdf, value))
	return c
}

func (c *Query) JoinInner(table string, link string) *Query{
	c.dataForBuilding = append(c.dataForBuilding, c.trf(JoinInner, table))
	c.dataForBuilding = append(c.dataForBuilding, c.trf(JoinLink, link))
	return c
}

func (c *Query) JoinLeft(table string, link string) *Query{
	c.dataForBuilding = append(c.dataForBuilding, c.trf(JoinLeft, table))
	c.dataForBuilding = append(c.dataForBuilding, c.trf(JoinLink, link))
	return c
}

func (c *Query) BktStart() *Query{
	c.dataForBuilding = append(c.dataForBuilding, c.trf(BktStart, true))
	return c
}

func (c *Query) BktEnd() *Query{
	c.dataForBuilding = append(c.dataForBuilding, c.trf(BktEnd, true))
	return c
}

func (c *Query) Update(table string, data ...map[string]interface{}) *Query {
	c.dataForBuilding = append(c.dataForBuilding, c.trf(Update, c.mysqlRealEscapeString(table)))
	var UpdateSetSlice []string
	for _, r := range data {
		for field, result := range r{
			UpdateSetSlice = append(UpdateSetSlice, fmt.Sprintf("%v='%v'", c.mysqlRealEscapeString(field), c.mysqlRealEscapeString(result)))
		}
	}
	c.dataForBuilding = append(c.dataForBuilding, c.trf(UpdateSet, strings.Join(UpdateSetSlice, ", ")))

	return c
}

func (c *Query) Incr(data ...interface{}) *Query{
	if len(data) == 1 {
		c.dataForBuilding = append(c.dataForBuilding, c.trf(Incr, c.mysqlRealEscapeString(data[0])))
	}else{
		for i, value := range data {
			if i==0 {
				c.dataForBuilding = append(c.dataForBuilding, c.trf(Update, c.mysqlRealEscapeString(value)))
			} else {
				c.dataForBuilding = append(c.dataForBuilding, c.trf(Incr, c.mysqlRealEscapeString(value)))
			}
		}
	}

	return c
}

func (c *Query) Decr(data ...interface{}) *Query{
	if len(data) == 1 {
		c.dataForBuilding = append(c.dataForBuilding, c.trf(Decr, c.mysqlRealEscapeString(data[0])))
	}else{
		for i, value := range data {
			if i==0 {
				c.dataForBuilding = append(c.dataForBuilding, c.trf(Update, c.mysqlRealEscapeString(value)))
			} else {
				c.dataForBuilding = append(c.dataForBuilding, c.trf(Decr, c.mysqlRealEscapeString(value)))
			}
		}
	}

	return c
}

func (c *Query) OrderBy(direction string, order ...string) *Query{
	sqlSlice := make([]string, 0)
	for _, value := range order {
		sqlSlice = append(sqlSlice, c.mysqlRealEscapeString(value))
	}
	c.dataForBuilding = append(c.dataForBuilding, c.trf(OrderBy, fmt.Sprintf("%s %s", strings.Join(sqlSlice, ", "), direction)))
	return c
}

func (c *Query) GroupBy(group ...string) *Query{
	for _, value := range group {
		c.dataForBuilding = append(c.dataForBuilding, c.trf(GroupBy, c.mysqlRealEscapeString(value)))
	}
	return c
}

func (c *Query) Limit(count string) *Query{
	c.dataForBuilding = append(c.dataForBuilding, c.trf(Limit, count))
	return c
}

func (c *Query) Offset(count string) *Query{
	c.dataForBuilding = append(c.dataForBuilding, c.trf(Offset, count))
	return c
}

func (c *Query) AddColumn(columns ...*Schema) *Query{
	sqlColumn := make([]string, 0, len(columns))
	var sqlPrimaryKey string
	sqlColumnIndex := make([]string, 0)

	for _, r := range columns {
		generateSql, primaryKey, IndexColumn := r.returnColumn()
		sqlColumn = append(sqlColumn, generateSql)
		if len(primaryKey)>0 {
			sqlPrimaryKey = primaryKey
		}
		if len(IndexColumn)>0{
			sqlColumnIndex = append(sqlColumnIndex, IndexColumn)
		}
	}

	if len(c.schemaColumns) != 0 && len(sqlColumn)>0 {
		c.schemaColumns = append(c.schemaColumns, sqlColumn...)
	}else if len(c.schemaColumns) == 0 && len(sqlColumn)>0{
		c.schemaColumns = make([]string, 0)
		c.schemaColumns = append(c.schemaColumns, sqlColumn...)
	}

	if len(sqlPrimaryKey)>0 {
		c.schemaPrimaryKey = sqlPrimaryKey
	}

	if len(c.schemaIndex) != 0 && len(sqlColumnIndex)>0 {
		c.schemaIndex = append(c.schemaIndex, sqlColumnIndex...)
	}else if len(c.schemaIndex) == 0 && len(sqlColumnIndex)>0{
		c.schemaIndex = make([]string, 0)
		c.schemaIndex = append(c.schemaIndex, sqlColumnIndex...)
	}


	return c
}

func (c *Query) AfterCreateTable(f func(*Query))*Query{
	c.schemaAfterCreateTable = f
	return c
}

func (c *Query) BeforeCreateTable(f func(*Query))*Query{
	c.schemaBeforeCreateTable = f
	return c
}

func (c *Query) AfterDeleteTable(f func(*Query))*Query{
	c.schemaAfterDeleteTable = f
	return c
}

func (c *Query) BeforeDeleteTable(f func(*Query))*Query{
	c.schemaBeforeDeleteTable = f
	return c
}

func (c *Query)TableEngine(name string)*Query{
	c.tableEngine = name
	return c
}

func (c *Query)CreateTable(name string){
	c.schemaTableName = name
	if c.schemaAfterCreateTable != nil {
		c.schemaAfterCreateTable(c)
	}

	sqlRequest := fmt.Sprintf("CREATE TABLE `%s` (%s)ENGINE=%s", name, c.getColumnsTable(), c.tableEngine)

	c.sqlRequest(sqlRequest)

	if c.schemaBeforeCreateTable != nil {
		c.schemaBeforeCreateTable(c)
	}

	c.flushSchema()
}
func (c *Query)CreateTableIfNotExist(name string){
	c.schemaTableName = name
	if c.schemaAfterCreateTable != nil {
		c.schemaAfterCreateTable(c)
	}

	sqlRequest := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (%s)ENGINE=%s", name, c.getColumnsTable(), c.tableEngine)

	c.sqlRequest(sqlRequest)

	if c.schemaBeforeCreateTable != nil {
		c.schemaBeforeCreateTable(c)
	}

	c.flushSchema()
}

func (c *Query)getColumnsTable()string{
	if len(c.tableEngine) == 0 {
		c.tableEngine = "INNODB"
	}
	sqlRequestBuilder := make([]string, 0)
	if len(c.schemaColumns)>0{
		sqlRequestBuilder = append(sqlRequestBuilder, c.schemaColumns...)
	}
	if len(c.schemaPrimaryKey)>0{
		sqlRequestBuilder = append(sqlRequestBuilder, fmt.Sprintf("PRIMARY KEY (`%s`)", c.schemaPrimaryKey))
	}
	if len(c.schemaIndex)>0{
		sqlRequestBuilder = append(sqlRequestBuilder, strings.Join(c.schemaIndex, ", "))
	}
	return strings.Join(sqlRequestBuilder, ", ")
}

func (c *Query)flushSchema(){
	c.schemaColumns = nil
	c.schemaIndex = nil
	c.scriptToConsole = false
}

func (c *Query)GetTable()string{
	return c.schemaTableName
}

func (c *Query)GetPrimaryKey()string{
	return c.schemaPrimaryKey
}

func (c *Query)DropTable(tables ...string){
	if c.schemaAfterDeleteTable != nil {
		c.schemaAfterDeleteTable(c)
	}

	sqlRequest := fmt.Sprintf("DROP TABLE `%s`", strings.Join(tables, "`, `"))

	c.sqlRequest(sqlRequest)

	if c.schemaBeforeDeleteTable != nil {
		c.schemaBeforeDeleteTable(c)
	}
}

func (c *Query)DropTableIfExists(tables ...string){
	if c.schemaAfterDeleteTable != nil {
		c.schemaAfterDeleteTable(c)
	}

	sqlRequest := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", strings.Join(tables, "`, `"))

	c.sqlRequest(sqlRequest)

	if c.schemaBeforeDeleteTable != nil {
		c.schemaBeforeDeleteTable(c)
	}
}

func (c *Query)ShowSqlInConsole()*Query{
	c.scriptToConsole = true
	return c
}

func (c *Query)sqlRequest(sqlRequest string){

	c.toSendConsole(sqlRequest)

	ins, err := Conn().Prepare(sqlRequest)
	if err != nil {
		panic(err.Error())
	}

	_, err = ins.Exec()

	if err != nil {
		panic(err.Error())
	}
}



func (c *Query) Build() string{
	sqlRequest := ""
	var SelectSlice       []string
	var FromSlice         []string
	var JoinSlice         []string
	var WhereSlice        []string
	var OrderBySlice =    make([]string, 0)
	var GroupBySlice =    make([]string, 0)
	var LimitSlice          string
	var OffsetSlice         string
	var UpdateSlice         string
	var UpdateParamsSlice []string


	for _, box := range c.dataForBuilding{
		for key, value := range box{
			switch key{
			case Select:
				SelectSlice = append(SelectSlice, fmt.Sprintf("%v", c.mysqlRealEscapeString(value)))

			case From:
				FromSlice = append(FromSlice, fmt.Sprintf("%v", value))

			case Field:
				var field string
				if value == nil{
				field = fmt.Sprintf("%v", c.mysqlRealEscapeString(value))
			} else{
				field = fmt.Sprintf("%v", c.mysqlRealEscapeString(value))
			}
				WhereSlice = append(WhereSlice, field)

			case BetweenValue1:
				WhereSlice = append(WhereSlice, fmt.Sprintf(" BETWEEN '%s'", c.mysqlRealEscapeString(value)))
			case BetweenValue2:
				WhereSlice = append(WhereSlice, fmt.Sprintf(" AND '%s'", c.mysqlRealEscapeString(value)), "AND")

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

			case NotIn:
				WhereSlice = append(WhereSlice, fmt.Sprintf("%v NOT IN ", c.mysqlRealEscapeString(value)))

			case InValue:
				WhereSlice = append(WhereSlice, fmt.Sprintf("%v", value))
				WhereSlice = append(WhereSlice, "AND")

			case InMdf:
				WhereSlice = append(WhereSlice, fmt.Sprintf("%v IN ", value))

			case NotInMdf:
				WhereSlice = append(WhereSlice, fmt.Sprintf("%v NOT IN ", value))

			case InValueMdf:
				WhereSlice = append(WhereSlice, fmt.Sprintf("(%v)", value))
				WhereSlice = append(WhereSlice, "AND")

			case JoinInner:
				JoinSlice = append(JoinSlice, fmt.Sprintf("INNER JOIN %v", value))

			case JoinLeft:
				JoinSlice = append(JoinSlice, fmt.Sprintf("LEFT JOIN %v", value))

			case JoinLink:
				JoinSlice = append(JoinSlice, fmt.Sprintf(" ON %v", value))

			case BktStart:
				WhereSlice = append(WhereSlice, "(")
			case BktEnd:
				WhereSlice = append(WhereSlice, ")")
			case Update:
				UpdateSlice = fmt.Sprintf("UPDATE %v", value)
			case UpdateSet:
				UpdateParamsSlice = append(UpdateParamsSlice, fmt.Sprintf(" %s ", value))
			case Incr:
				UpdateParamsSlice = append(UpdateParamsSlice, fmt.Sprintf(" %s = %s+1 ", value, value))
			case Decr:
				UpdateParamsSlice = append(UpdateParamsSlice, fmt.Sprintf(" %s = %s-1 ", value, value))
			case OrderBy:
				OrderBySlice = append(OrderBySlice, fmt.Sprintf("%v", value))
			case GroupBy:
				GroupBySlice = append(GroupBySlice, fmt.Sprintf("%v", value))
			case Limit:
				LimitSlice = fmt.Sprintf("%v", value)
			case Offset:
				OffsetSlice = fmt.Sprintf("%v", value)
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

	if len(GroupBySlice)>0 {
		sqlRequest += fmt.Sprintf(" GROUP BY %s", strings.Join(GroupBySlice, ", "))
	}

	if len(OrderBySlice)>0 {
		sqlRequest += fmt.Sprintf(" ORDER BY %s", strings.Join(OrderBySlice, ", "))
	}

	if len(LimitSlice)>0 {
		sqlRequest += fmt.Sprintf(" LIMIT %s", LimitSlice)
	}

	if len(OffsetSlice)>0 {
		sqlRequest += fmt.Sprintf(" OFFSET %s", OffsetSlice)
	}

	if len(c.alias)>0 {
		sqlRequest = fmt.Sprintf("(%s) AS %s", sqlRequest, c.alias)
	}

	return sqlRequest
}

func (c *Query) Apply()int64{
	var aff int64
	if textSql := c.Build(); textSql!=""{

		c.toSendConsole(textSql)

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

		c.toSendConsole(textSql)

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

		c.toSendConsole(textSql)

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


		c.toSendConsole(textSql)

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

//Добавление ключа и значения в общую корробку
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

func (c *Query) toSendConsole(sqlText string){
	if c.scriptToConsole{
		fmt.Println(sqlText)
	}
}
