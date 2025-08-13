package config

import (
	"errors"
	"os"

	"github.com/spf13/viper"
)

var ErrConfigLoadFailed = errors.New("failed to load config")

type Config struct {
	Host                 string `mapstructure:"HOST"`
	DbUrl                string `mapstructure:"DB_URL"`
	AccessTokenLifeSpan  int    `mapstructure:"ACCESS_TOKEN_LIFE_SPAN"`
	RefreshTokenLifeSpan int    `mapstructure:"REFRESH_TOKEN_LIFE_SPAN"`
	JwtAccessSecret      string `mapstructure:"JWT_ACCESS_SECRET"`
	JwtRefreshSecret     string `mapstructure:"JWT_REFRESH_SECRET"`
}

var globalConfig *Config

func loadConfig(path string) (config *Config, err error) {
	viper.AddConfigPath(path)

	if os.Getenv("GO_ENV") == "dev" {
		viper.SetConfigName("dev")
	} else {
		viper.SetConfigName("prod")
	}
	viper.SetConfigType("env")

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
		globalConfig, err = loadConfig(".")
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
