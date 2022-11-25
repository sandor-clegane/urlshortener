package app

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"

	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type Handler struct {
	*chi.Mux
	cfg     Config
	storage Storage
}

func NewHandler() *Handler {
	h := &Handler{
		Mux: chi.NewRouter(),
	}

	h.ConfigureHandler()
	h.InitHandler()

	return h
}

func (h *Handler) ConfigureHandler() {
	var c2 Config
	//parsing env config
	err := env.Parse(&h.cfg)
	if err != nil {
		log.Fatal(err)
	}

	//parsing command line config
	if !flag.Parsed() {
		flag.StringVar(&c2.ServerAddress, "a",
			DefaultServerAddress, "http server launching address")
		flag.StringVar(&c2.BaseURL, "b", DefaultBaseURL,
			"base address of resulting shortened URL")
		flag.StringVar(&c2.FileStoragePath, "f", DefaultFileStoragePath,
			"path to file with shortened URL")
		flag.Parse()
	}

	h.cfg.ApplyConfig(c2)
}

func (h *Handler) InitHandler() {
	//init storage
	h.InitStorage()
	//push middlewares
	h.Use(gzipHandle)
	h.Use(ungzipHandle)
	//configuration handlers
	h.MethodFunc("GET", "/{id}", h.getHandler)
	h.MethodFunc("POST", "/", h.postHandler)
	h.MethodFunc("POST", "/api/shorten", h.postHandlerJSON)
}

func (h *Handler) InitStorage() {
	if h.cfg.FileStoragePath == "" {
		h.storage = NewInMemoryStorage()
	} else {
		h.storage = NewFileStorage(h.cfg.FileStoragePath)
	}
}

func (h *Handler) Start() error {
	return http.ListenAndServe(h.cfg.ServerAddress, h)
}

//пишем свой джоин потому что проект на версии go 1.16
func Join(basePath string, paths ...string) (*url.URL, error) {
	u, err := url.Parse(basePath)
	if err != nil {
		return nil, fmt.Errorf("invalid url")
	}

	p2 := append([]string{u.Path}, paths...)
	result := path.Join(p2...)
	u.Path = result

	return u, nil
}

func (h *Handler) shortenURL(_ *url.URL) (*url.URL, error) {
	return Join(h.cfg.BaseURL, uuid.NewString())
}
