package main

import (
	"builder/example/db"
	"fmt"
)

func main(){
	//Example 1
	query := db.BuilderQuery()
	data := query.
		Select("*").
		From("test_data").
		Where("data_id", "1").
		Rows()
	fmt.Println(data)

	//Example 2
	query = db.BuilderQuery()
	var fields []interface{}
	fields = append(fields, "217")
	fields = append(fields, "216")
	fields = append(fields, "215")
	fields = append(fields, "214")
	fields = append(fields, "213")
	fields = append(fields, "212")

	data = query.
			Select("*").
			From(db.BuilderQuery().Alias("a").Select("*").From("test_data").WhereIn("data_id", fields).Build()).
			JoinInner("test_data as t", "t.data_id = a.data_id").
			Where("a.data_id", "212").
			Rows()

	fmt.Println(data)

	query.AddColumn(
		query.ColumnPrimaryKey(11).Name("id_column"),
		query.ColumnInt(11).Name("count_column").NotNull(true).DefaultValue(0),
		query.ColumnString(250).Name("string_column").Comment("Comment"),
		)
}
