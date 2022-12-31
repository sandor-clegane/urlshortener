package app

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sandor-clegane/urlshortener/internal/config"
	"github.com/sandor-clegane/urlshortener/internal/handlers/url"
	router2 "github.com/sandor-clegane/urlshortener/internal/router"
	"github.com/sandor-clegane/urlshortener/internal/storages"
)

const (
	readTimeout    = 10 * time.Second
	writeTimeout   = 10 * time.Second
	maxHeaderBytes = 1 << 20
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
		Addr:           cfg.ServerAddress,
		Handler:        router,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}
	defer listenToShutdownSignal(server, stg)
	return &App{
		server: server,
	}, nil
}

func listenToShutdownSignal(server *http.Server, storage storages.Storage) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		server.Close()
		storage.Stop()
	}()
}

func (app *App) Run() error {
	return app.server.ListenAndServe()
}
