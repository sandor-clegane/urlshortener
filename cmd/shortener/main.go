package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/sandor-clegane/urlshortener/internal/app"
	"log"
)

var (
	configPath string
)

func init() {
	flag.StringVar(
		&configPath,
		"config-path",
		"config.toml",
		"path to config file",
	)
}

func main() {
	flag.Parse()

	config := app.NewConfig()
	_, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		log.Fatal(err)
	}

	s := app.New(config)

	log.Fatal(s.Start())
}
