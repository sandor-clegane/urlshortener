package main

import (
	"log"
	"net/http"

	"github.com/sandor-clegane/urlshortener/internal/app"
)

func main() {
	h, err := app.New()
	if err != nil {
		log.Fatal(err)
	}
	if err = h.Run(); err != http.ErrServerClosed && err != nil {
		log.Fatal(err)
	}
}
