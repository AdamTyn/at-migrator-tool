package main

import (
	"at-migrator-tool/internal"
	col "at-migrator-tool/internal/collector"
	"at-migrator-tool/internal/conf"
	"at-migrator-tool/internal/pkg"
	"at-migrator-tool/internal/process"
	_ "embed"
	"encoding/json"
)

//go:embed config.json
var configByte []byte

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

	badDataCollector := col.NewBadDataCollector(500, rd)
	fsRobotCollector := col.NewFsRobotCollector(3, appC.Webhook)

	operateRecordMigrator := process.NewOperateRecordMigrator(app, pg1, rd)
	operateRecordMigrator.
		Add(col.NewCompanyOLogCollector(5000, pg2)).
		Add(col.NewDeliveryOLogCollector(5000, pg2)).
		Add(col.NewInternOLogCollector(5000, pg2)).
		Add(badDataCollector).
		Add(fsRobotCollector)
	app.Add(operateRecordMigrator)

	deliverUncheckMigrator := process.NewDeliverUncheckMigrator(app, pg1, pg1, rd)
	deliverUncheckMigrator.
		Add(col.NewDeliverUncheckCollector(8000, pg1)).
		Add(badDataCollector).
		Add(fsRobotCollector)
	app.Add(deliverUncheckMigrator)

	app.Go()
}

func getAppConfig() *conf.App {
	var appC conf.App
	err := json.Unmarshal(configByte, &appC)
	if err != nil {
		panic(err)
	}
	return &appC
}
