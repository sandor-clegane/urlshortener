package main

import (
	"log"
	"net/http"

	"github.com/sandor-clegane/urlshortener/internal/app"
	"github.com/sandor-clegane/urlshortener/internal/config"
)

func main() {
	var cfg config.Config
	cfg.Init()
	h, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if err = h.Run(); err != http.ErrServerClosed && err != nil {
		log.Fatal(err)
	}
}
