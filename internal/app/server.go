package app

import (
	"net/http"

	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi"
	"github.com/sandor-clegane/urlshortener/internal/config"
	"github.com/sandor-clegane/urlshortener/internal/handlers/db"
	"github.com/sandor-clegane/urlshortener/internal/handlers/url"
	"github.com/sandor-clegane/urlshortener/internal/storages"
)

type App struct {
	*chi.Mux
	Cfg  config.Config
	dbh  db.DBHandler
	urlh url.URLHandler
}

func New() (*App, error) {
	h := &App{
		Mux: chi.NewRouter(),
	}

	err := h.initConfig()
	if err != nil {
		return nil, err
	}

	err = h.initHandlers()
	if err != nil {
		return nil, err
	}

	return h, nil
}

func (h *App) initConfig() error {
	var c2 config.Config
	//parsing env config
	err := env.Parse(&h.Cfg)
	if err != nil {
		return err
	}
	//parsing command line config
	c2.ParseArgsCMD()
	//applying config
	h.Cfg.ApplyConfig(c2)
	return nil
}

//TODO паттерны стоит вынести в константы
func (h *App) initHandlers() error {
	stg, err := storages.CreateStorage(h.Cfg)
	if err != nil {
		return err
	}
	h.dbh, err = db.NewDBHandler(h.Cfg.DatabaseDSN)
	if err != nil {
		return err
	}
	h.urlh = url.New(stg, h.Cfg)

	h.Use(GzipCompressHandle, GzipDecompressHandle, h.urlh.GetAuthorizationMiddleware())

	h.Post("/", h.urlh.ShortenURL)
	h.Post("/api/shorten", h.urlh.ShortenURLwJSON)
	h.Post("/api/shorten/batch", h.urlh.ShortenSomeURL)

	h.Get("/ping", h.dbh.PingConnectionDB)
	h.Get("/{id}", h.urlh.ExpandURL)
	h.Get("/api/user/urls", h.urlh.GetAllURL)
	return nil
}

func (h *App) Run() error {
	return http.ListenAndServe(h.Cfg.ServerAddress, h)
}
