package pkg

import (
	"at-migrator-tool/internal/conf"
	"bytes"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func DB(c *conf.Data_Database) (db *sql.DB) {
	switch c.Driver {
	case "postgres":
		db = NewPostgreSQL(c)
	case "mysql":
		db = NewMySQL(c)
	default:
		panic("unknown database driver")
	}
	return db
}

func JsonPost(url string, input interface{}) ([]byte, error) {
	data, err0 := json.Marshal(input)
	if err0 != nil {
		return nil, err0
	}
	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Add("Content-Type", "application/json;charset=UTF-8")
	resp, err1 := client.Do(req)
	if err1 != nil {
		return nil, err1
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
