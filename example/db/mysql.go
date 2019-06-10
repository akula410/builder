package db

import (
	c "github.com/akula410/connect"
)

var MySql c.MySql
func init(){
	MySql.DBName = "golang"
	MySql.Host = "localhost"
	MySql.User = "root"
	MySql.Password = ""
	MySql.Port = "3306"
	MySql.Charset = "utf8"
	MySql.InterpolateParams = true
	MySql.MaxOpenCoons = 10
}
