package app

const (
	DefaultServerAddress   = "localhost:8080"
	DefaultBaseURL         = "http://localhost:8080/"
	DefaultFileStoragePath = ""
	DefaultKey             = "SuperSecretKey2022"
	DefaultDatabaseDSN     = ""
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL"       envDefault:"http://localhost:8080/"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:""`
	Key             string `env:"SECRET_KEY" envDefault:"SuperSecretKey2022"`
	DatabaseDSN     string `env:"DATABASE_DSN" envDefault:""`
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
