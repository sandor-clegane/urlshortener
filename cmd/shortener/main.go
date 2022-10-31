package main

import (
	"github.com/sandor-clegane/urlshortener/internal/app"
	"log"
)

func main() {
	s := app.New()
	log.Fatal(s.Start())
}
