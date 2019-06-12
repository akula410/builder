package db

import(
	"database/sql"
	"github.com/akula410/builder"
)

type query struct {
	builder.Query
}

func init(){
	builder.Conn =  func() *sql.DB{
		return MySql.Connect()
	}

	builder.ConnClose = func(){
		MySql.Close()
	}
}

func BuilderQuery() *query{
	db := &query{}
	return db
}
