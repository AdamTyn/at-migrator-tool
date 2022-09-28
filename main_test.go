package main

import (
	"at-migrator-tool/internal"
	col "at-migrator-tool/internal/collector"
	"at-migrator-tool/internal/process"
	"testing"
	"time"
)

func TestGetAppConfig(t *testing.T) {
	appC := getAppConfig()
	t.Log(appC)
	t.Log(appC.Migrator.OperateRecord.FetchStep)
	t.Log(appC.Migrator.OperateRecord.MaxEmptyFetchNum)
	t.Log(appC.Webhook.FsRobot)
}

func TestNoProcess(t *testing.T) {
	appC := getAppConfig()
	app := internal.NewApp(appC)
	app.Go()
}

func TestRunHttp(t *testing.T) {
	appC := getAppConfig()
	app := internal.NewApp(appC)
	http := process.NewHttp(app)
	app.Add(http).
		Go()
}

func TestFsRobotCollector(t *testing.T) {
	appC := getAppConfig()
	c := col.NewFsRobotCollector(3, appC.Webhook)
	go c.Listen()
	m := make(map[string]interface{})
	m["msg_type"] = "text"
	m["content"] = map[string]string{
		"text": "[at-migrator-tool] func TestFsRobotCollector(t *testing.T)",
	}
	err := c.Put(m)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Second)
}
