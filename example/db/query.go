package db

import(
	"database/sql"
	"github.com/akula410/builderQuery"
)

type query struct {
	builderQuery.Query
}

func init(){
	builderQuery.Conn =  func() *sql.DB{
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
