package main

import (
	"log"
	"net/http"

	"github.com/sandor-clegane/urlshortener/internal/app"
)

func main() {
	h := app.New()
	if err := h.Run(); err != http.ErrServerClosed && err != nil {
		log.Fatal(err)
	}
}
