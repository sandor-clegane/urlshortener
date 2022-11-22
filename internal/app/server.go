package app

import (
	"flag"
	"fmt"
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

//TODO передаваемые параметры не валидируются
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL"       envDefault:"http://localhost:8080/"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:""`
}

func (c *Config) isDefault() bool {
	return c.BaseURL == "http://localhost:8080/" &&
		c.ServerAddress == "localhost:8080" && c.FileStoragePath == ""
}

//TODO обработать ошибки при создании
func NewHandler() *Handler {
	h := &Handler{
		Mux: chi.NewRouter(),
	}
	h.ConfigureHandler()
	h.InitHandler()

	return h
}

func (h *Handler) ConfigureHandler() {
	//parsing command line config
	if !flag.Parsed() {
		flag.StringVar(&h.cfg.ServerAddress, "a",
			"localhost:8080", "http server launching address")
		flag.StringVar(&h.cfg.BaseURL, "b", "http://localhost:8080/",
			"base address of resulting shortened URL")
		flag.StringVar(&h.cfg.FileStoragePath, "f", "",
			"path to file with shortened URL")
		flag.Parse()
	}
	//parsing env config
	if h.cfg.isDefault() {
		_ = env.Parse(&h.cfg)
	}
}

func (h *Handler) InitHandler() {
	//init storage
	h.InitStorage()
	//configuration handlers
	h.MethodFunc("GET", "/{id}", h.getHandler)
	h.MethodFunc("POST", "/", h.postHandler)
	h.MethodFunc("POST", "/api/shorten", h.postHandlerJSON)
}

//TODO мне не нравится эта функция ,возможно стоит по другому создавать хранилище
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
