package config

import (
	"errors"
	"os"

	"github.com/spf13/viper"
)

var ErrConfigLoadFailed = errors.New("failed to load config")

type Config struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Db   struct {
		Host      string `mapstructure:"host"`
		Port      int    `mapstructure:"port"`
		User      string `mapstructure:"user"`
		Password  string `mapstructure:"password"`
		Database  string `mapstructure:"database"`
		EnableSSL bool   `mapstructure:"enable_ssl"`
	} `mapstructure:"db"`
	SSL struct {
		Enabled  bool   `mapstructure:"enabled"`
		CertFile string `mapstructure:"cert_file"`
		KeyFile  string `mapstructure:"key_file"`
	} `mapstructure:"ssl"`
	JWT struct {
		Lifespan struct {
			AccessToken  int `mapstructure:"access_token"`
			RefreshToken int `mapstructure:"refresh_token"`
		} `mapstructure:"lifespan"`
		Secrets struct {
			Access  string `mapstructure:"access"`
			Refresh string `mapstructure:"refresh"`
		} `mapstructure:"secrets"`
	} `mapstructure:"jwt"`
}

var globalConfig *Config

func loadConfig() (config *Config, err error) {
	if os.Getenv("GO_ENV") == "dev" {
		viper.SetConfigName("dev-config")
		viper.AddConfigPath(".")
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath("$HOME/.config/go-todo")
	}
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	var localConfig Config

	err = viper.Unmarshal(&localConfig)
	config = &localConfig
	return
}

func Get() (config *Config, err error) {
	if globalConfig == nil {
		globalConfig, err = loadConfig()
		if err != nil {
			err = errors.Join(ErrConfigLoadFailed, err)
			return
		}
	}
	config = globalConfig
	return
}

// Returns the value of GO_ENV environment variable.
func GetGoEnv() string {
	return os.Getenv("GO_ENV")
}
