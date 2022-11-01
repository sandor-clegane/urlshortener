package main

import (
	"github.com/sandor-clegane/urlshortener/internal/app"
	"log"
	"net/http"
)

func main() {
	h := app.NewHandler()
	log.Fatal(http.ListenAndServe(":8080", h))
}
