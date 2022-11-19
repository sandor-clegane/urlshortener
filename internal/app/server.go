package app

import (
	"net/url"
	"sync"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type Handler struct {
	*chi.Mux
	storage map[string]string
	lock    sync.RWMutex
}

func NewHandler() *Handler {
	//creating handler
	h := &Handler{
		Mux:     chi.NewRouter(),
		storage: make(map[string]string),
	}
	//configuration router
	h.MethodFunc("GET", "/{id}", h.getHandler)
	h.MethodFunc("POST", "/", h.postHandler)
	h.MethodFunc("POST", "/api/shorten", h.postHandlerJSON)

	return h
}

func shortenURL(_ *url.URL) url.URL {
	return url.URL{
		Scheme: "http",
		Host:   "localhost:8080",
		Path:   uuid.New().String(),
	}
}
