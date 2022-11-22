package main

import (
	"log"
	"net/http"

	"github.com/sandor-clegane/urlshortener/internal/app"
)

func main() {
	h := app.NewHandler()
	h.ConfigureHandler()
	h.InitHandler()
	if err := h.Start(); err != http.ErrServerClosed && err != nil {
		log.Fatal(err)
	}

}
