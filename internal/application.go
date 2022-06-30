package internal

import (
	"at-migrator-tool/internal/conf"
	"at-migrator-tool/internal/contract"
	"log"
	"os"
	"os/signal"
	"syscall"
)

/**
 * Application
 * @author ctc
 * @description app容器
 */
type Application struct {
	Conf      *conf.App
	Logger    *log.Logger
	processes map[string]contract.Process
}

func NewApp(c *conf.App) *Application {
	app := &Application{
		Logger:    log.Default(),
		Conf:      c,
		processes: make(map[string]contract.Process),
	}
	return app
}

func (app Application) Log(msg string) {
	app.Logger.Println(msg)
}

func (app Application) Logf(format string, v ...interface{}) {
	app.Logger.Printf(format, v)
}

/**
 * Application.Add
 * @author ctc
 * @description 添加Process到app容器（启用CPU多核后Process将会并行）
 */
func (app *Application) Add(p contract.Process) *Application {
	app.processes[p.Name()] = p
	return app
}

/**
 * Application.Go
 * @author ctc
 * @description 启动app容器
 */
func (app *Application) Go() {
	app.Logf("[Info] app %s start", app.Conf.Name)
	if app.processes == nil || len(app.processes) < 1 {
		app.Log("[Notice] no process to run, bye bye")
		return
	}
	defer func() {
		if err := recover(); err != nil {
			app.Logf("[Error] %s", err)
		}
	}()
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	for k := range app.processes {
		go app.processes[k].Start()
	}
	for {
		select {
		case s := <-c:
			app.Logf("[Notice] get signal %s, app ending...", s)
			app.Stop()
			close(c)
			return
		}
	}
}

/**
 * Application.Stop
 * @author ctc
 * @description 平滑终止app容器
 */
func (app *Application) Stop() {
	app.Log("[Notice] app ending")
	for k := range app.processes {
		app.processes[k].Shutdown()
	}
	app.Log("[Notice] app ended")
}
