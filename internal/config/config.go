package config

import "flag"

const (
	DefaultServerAddress   = "localhost:8080"
	DefaultBaseURL         = "http://localhost:8080/"
	DefaultFileStoragePath = ""
	DefaultKey             = "SuperSecretKey2022"
	DefaultDatabaseDSN     = "user=pqgotest dbname=pqgotest sslmode=verify-full"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL"       envDefault:"http://localhost:8080/"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:""`
	Key             string `env:"SECRET_KEY" envDefault:"SuperSecretKey2022"`
	DatabaseDSN     string `env:"DATABASE_DSN" envDefault:"user=pqgotest dbname=pqgotest sslmode=verify-full"`
}

func (c *Config) ParseArgsCMD() {
	if !flag.Parsed() {
		flag.StringVar(&c.ServerAddress, "a",
			DefaultServerAddress, "http server launching address")
		flag.StringVar(&c.BaseURL, "b", DefaultBaseURL,
			"base address of resulting shortened URL")
		flag.StringVar(&c.FileStoragePath, "f", DefaultFileStoragePath,
			"path to file with shortened URL")
		flag.StringVar(&c.DatabaseDSN, "d", DefaultDatabaseDSN,
			"DB connection address")
		flag.Parse()
	}
}

func (c *Config) ApplyConfig(other Config) {
	if c.ServerAddress == DefaultServerAddress {
		c.ServerAddress = other.ServerAddress
	}
	if c.BaseURL == DefaultBaseURL {
		c.BaseURL = other.BaseURL
	}
	if c.FileStoragePath == DefaultFileStoragePath {
		c.FileStoragePath = other.FileStoragePath
	}
	if c.Key == DefaultKey {
		c.Key = other.Key
	}
	if c.DatabaseDSN == DefaultDatabaseDSN {
		c.DatabaseDSN = other.DatabaseDSN
	}
}
