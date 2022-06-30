package main

import (
	"at-migrator-tool/internal"
	col "at-migrator-tool/internal/collector"
	"at-migrator-tool/internal/conf"
	"at-migrator-tool/internal/pkg"
	"at-migrator-tool/internal/process"
	"encoding/json"
	"io/ioutil"
	"os"
)

func main() {
	appC := getAppConfig()
	app := internal.NewApp(appC)

	http := process.NewHttp(app)
	app.Add(http)

	pg1 := pkg.DB(appC.Data.Source) // 迁移前的数据源
	pg2 := pkg.DB(appC.Data.Target) // 迁移后的数据库
	rd := pkg.NewRedis(appC.Data.Redis)
	defer func() {
		_ = pg1.Close()
		_ = pg2.Close()
		_ = rd.Close()
	}()

	operateRecordMigrator := process.NewOperateRecordMigrator(app, pg1, rd)
	operateRecordMigrator.
		Add(col.NewCompanyOLogCollector(1000, pg2, app.Logger)).
		Add(col.NewDeliveryOLogCollector(5000, pg2, app.Logger)).
		Add(col.NewInternOLogCollector(5000, pg2, app.Logger)).
		Add(col.NewBadDataCollector(500, rd, app.Logger)).
		Add(col.NewFsRobotCollector(3, appC.Webhook, app.Logger))
	app.Add(operateRecordMigrator).
		Go()
}

func getAppConfig() *conf.App {
	fb, err := os.Open("../config.json")
	if err != nil {
		panic(err)
	}
	defer fb.Close()
	var bytes []byte
	bytes, err = ioutil.ReadAll(fb)
	if err != nil {
		panic(err)
	}
	var appC conf.App
	err = json.Unmarshal(bytes, &appC)
	if err != nil {
		panic(err)
	}
	return &appC
}
