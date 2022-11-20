package app

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sync"

	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type Handler struct {
	cfg Config
	*chi.Mux
	storage map[string]string
	lock    sync.RWMutex
}

//TODO передаваемые параметры не валидируются
type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL       string `env:"BASE_URL"       envDefault:"http://localhost:8080/"`
}

//TODO обработать ошибки при создании
func NewHandler() *Handler {
	//creating handler
	h := &Handler{
		Mux:     chi.NewRouter(),
		storage: make(map[string]string),
	}
	//parsing config
	_ = env.Parse(&h.cfg)
	//configuration handlers
	h.MethodFunc("GET", "/{id}", h.getHandler)
	h.MethodFunc("POST", "/", h.postHandler)
	h.MethodFunc("POST", "/api/shorten", h.postHandlerJSON)

	return h
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
