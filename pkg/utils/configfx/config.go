package configfx

import (
	"os"

	"go.uber.org/fx"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Log LogConfig `yaml:"log"`
	DB  DBConfig  `yaml:"db"`
}

type LogConfig struct {
	Level string `yaml:"level"`
	Path  string `yaml:"path"`
}

type DBConfig struct {
	DSN         string `yaml:"dsn"`
	WALMode     bool   `yaml:"wal_mode"`
	BusyTimeout int    `yaml:"busy_timeout"`
}

func DefaultConfig() Config {
	return Config{
		Log: LogConfig{Level: "info"},
		DB: DBConfig{
			DSN:         "file:data.db",
			WALMode:     true,
			BusyTimeout: 5000,
		},
	}
}

func loadConfig(path string) (Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return Config{}, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func ProvideConfig() (Config, error) {
	return loadConfig("config.yaml")
}

var Module = fx.Module("configfx",
	fx.Provide(ProvideConfig),
)
