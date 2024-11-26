package config

import (
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env                    Env                          `env:"ENV" env-required:"true"`
	LogLevel               string                       `env:"LOG_LEVEL" env-default:"warn"`
	DBConfig               DBConfig                     `env-prefix:"DB_"`
	HTTPServer             HTTPServerConfig             `env-prefix:"HTTP_SERVER_"`
	SongInfoIntegrationAPI SongInfoIntegrationAPIConfig `env-prefix:"SONG_INFO_INTEGRATION_API_"`
}

type Env string

const (
	EnvLocal Env = "local"
	EnvProd  Env = "prod"
)

type HTTPServerConfig struct {
	Host    string        `env:"HOST" env-required:"true"`
	Port    string        `env:"PORT" env-required:"true"`
	Timeout time.Duration `env:"TIMEOUT" env-default:"4s"`
}

type DBConfig struct {
	Host     string `env:"HOST" env-required:"true"`
	Port     string `env:"PORT" env-required:"true"`
	DBName   string `env:"NAME" env-required:"true"`
	SSLMode  string `env:"SSL_MODE" env-required:"true"`
	Username string `env:"USERNAME" env-required:"true"`
	Password string `env:"PASSWORD" env-required:"true"`
}

type SongInfoIntegrationAPIConfig struct {
	Scheme       string `env:"SCHEME" env-required:"true"`
	Domain       string `env:"DOMAIN" env-required:"true"`
	SongInfoPath string `env:"SONG_INFO_PATH" env-required:"true"`
}

var (
	once sync.Once
	cfg  Config
)

func New() (Config, error) {
	var err error
	once.Do(func() {
		err = cleanenv.ReadEnv(&cfg)
	})

	return cfg, err
}
