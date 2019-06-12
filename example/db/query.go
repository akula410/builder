package db

import(
	"database/sql"
	"github.com/akula410/builder"
)

type query struct {
	builder.query
}

func init(){
	builder.Conn =  func() *sql.DB{
		return MySql.Connect()
	}

	builderQuery.ConnClose = func(){
		MySql.Close()
	}
}

func BuilderQuery() *query{
	db := &query{}
	return db
}
