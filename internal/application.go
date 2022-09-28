package internal

import (
	"at-migrator-tool/internal/conf"
	"at-migrator-tool/internal/contract"
	"at-migrator-tool/internal/pkg/log"
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
	processes map[string]contract.Process
}

func NewApp(c *conf.App) *Application {
	app := &Application{
		Conf:      c,
		processes: make(map[string]contract.Process),
	}
	return app
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
	log.InfoF("app [%s] start", app.Conf.Name)
	if app.processes == nil || len(app.processes) < 1 {
		log.Notice("no process to run, bye bye!")
		return
	}
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
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
			log.NoticeF("get signal [%s], app ending...", s)
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
	log.Notice("app ending")
	for k := range app.processes {
		app.processes[k].Shutdown()
	}
	log.Notice("app ended")
}
