package app

import (
	"net/http"
	"time"

	"github.com/sandor-clegane/urlshortener/internal/config"
	"github.com/sandor-clegane/urlshortener/internal/handlers/url"
	router2 "github.com/sandor-clegane/urlshortener/internal/router"
	"github.com/sandor-clegane/urlshortener/internal/storages"
)

const (
	rTimeout = 10 * time.Second
	wTimeout = 10 * time.Second
)

type App struct {
	server *http.Server
}

func New(cfg config.Config) (*App, error) {
	stg, err := storages.CreateStorage(cfg)
	if err != nil {
		return nil, err
	}
	router := router2.NewRouter(url.New(stg, cfg))
	server := &http.Server{
		Addr:         cfg.ServerAddress,
		Handler:      router,
		ReadTimeout:  rTimeout,
		WriteTimeout: wTimeout,
	}
	//defer closeHTTPServerAndStopWorkerPool(server, urlRepository)
	return &App{
		server: server,
	}, nil
}

func (h *App) Run() error {
	return h.server.ListenAndServe()
}
