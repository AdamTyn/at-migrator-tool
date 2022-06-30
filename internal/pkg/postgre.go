package pkg

import (
	"at-migrator-tool/internal/conf"
	"database/sql"
	_ "github.com/lib/pq"
)

func NewPostgreSQL(c *conf.Data_Database) *sql.DB {
	pg, err := sql.Open(c.Driver, c.Dsn)
	if err != nil {
		panic(err)
	}
	err = pg.Ping()
	if err != nil {
		panic(err)
	}
	return pg
}
