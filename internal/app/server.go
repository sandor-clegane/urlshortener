package app

import (
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"net/url"
)

type Handler struct {
	*chi.Mux
	storage map[string]string
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

	return h
}

func shortenURL(u *url.URL) url.URL {
	return url.URL{
		Scheme: "http",
		Host:   "localhost:8080",
		Path:   uuid.New().String(),
	}
}
