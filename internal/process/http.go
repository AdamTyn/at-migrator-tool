package process

import (
	"at-migrator-tool/internal"
	"at-migrator-tool/internal/pkg/log"
	"context"
	"net/http"
	"time"
)

const HttpProcess = "http-server"

type Http struct {
	server *http.Server
	app    *internal.Application
}

func NewHttp(app *internal.Application) *Http {
	return &Http{app: app}
}

func (h *Http) Name() string {
	return HttpProcess
}

func (h *Http) Start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK."))
	})
	h.server = &http.Server{
		Addr:         h.app.Conf.Server.Http.Endpoint,
		WriteTimeout: time.Second * time.Duration(h.app.Conf.Server.Http.Timeout),
		Handler:      mux,
	}
	log.InfoF("start http server listen: [%s]", h.app.Conf.Server.Http.Endpoint)
	if err := h.server.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			panic(err)
		}
	}
}

func (h *Http) Shutdown() {
	_ = h.server.Shutdown(context.Background())
	log.NoticeF("process [%s] shutdown", HttpProcess)
}
