package app

import (
	"log"
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

func New() *App {
	h := &App{
		Mux: chi.NewRouter(),
	}

	h.getConfig()
	h.initHandlers()

	return h
}

func (h *App) getConfig() {
	var c2 config.Config
	//parsing env config
	err := env.Parse(&h.Cfg)
	if err != nil {
		log.Fatal(err)
	}
	//parsing command line config
	c2.ParseArgsCMD()
	//applying config
	h.Cfg.ApplyConfig(c2)
}

//TODO паттерны стоит вынести в константы
func (h *App) initHandlers() {
	//init storage
	stg := storages.CreateStorage(h.Cfg)
	h.dbh = db.NewDBHandler(h.Cfg.DatabaseDSN)
	h.urlh = url.New(stg, h.Cfg)

	h.Use(GzipCompressHandle, GzipDecompressHandle, h.urlh.GetAuthorizationMiddleware())

	h.MethodFunc("POST", "/", h.urlh.ShortenURL)
	h.MethodFunc("POST", "/api/shorten", h.urlh.ShortenURLwJSON)
	h.MethodFunc("POST", "/api/shorten/batch", h.urlh.ShortenSomeURL)

	h.MethodFunc("GET", "/ping", h.dbh.PingConnectionDB)
	h.MethodFunc("GET", "/{id}", h.urlh.ExpandURL)
	h.MethodFunc("GET", "/api/user/urls", h.urlh.GetAllURL)
}

func (h *App) Start() error {
	return http.ListenAndServe(h.Cfg.ServerAddress, h)
}
