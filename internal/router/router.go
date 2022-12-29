package router

import (
	"github.com/go-chi/chi"
	"github.com/sandor-clegane/urlshortener/internal/handlers/url"
)

func NewRouter(h url.URLHandler) chi.Router {
	r := chi.NewRouter()
	r.Use(GzipCompressHandle, GzipDecompressHandle, h.GetAuthorizationMiddleware())
	r.Post("/", h.ShortenURL)
	r.Post("/api/shorten", h.ShortenURLwJSON)
	r.Post("/api/shorten/batch", h.ShortenSomeURL)
	r.Get("/ping", h.PingConnectionDB)
	r.Get("/{id}", h.ExpandURL)
	r.Get("/api/user/urls", h.GetAllURL)
	r.Delete("/api/user/urls", h.DeleteSomeURL)
	return r
}
