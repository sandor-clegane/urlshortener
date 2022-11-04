package main

import (
	"log"
	"net/http"

	"github.com/sandor-clegane/urlshortener/internal/app"
)

func main() {
	h := app.NewHandler()
	if err := http.ListenAndServe(":8080", h); err != http.ErrServerClosed {
		log.Fatal(err)
	}

}
