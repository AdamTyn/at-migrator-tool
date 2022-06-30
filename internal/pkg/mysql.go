package pkg

import (
	"at-migrator-tool/internal/conf"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

func NewMySQL(c *conf.Data_Database) *sql.DB {
	mysql, err := sql.Open(c.Driver, c.Dsn)
	if err != nil {
		panic(err)
	}
	err = mysql.Ping()
	if err != nil {
		panic(err)
	}
	return mysql
}
