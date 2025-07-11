package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
}

type ServerConfig struct {
	Mode         string `mapstructure:"mode"`
	Port         string `mapstructure:"port"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	RateLimit    int    `mapstructure:"rate_limit"`
}

func Setup() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.read_timeout", 10)
	viper.SetDefault("server.write_timeout", 60)
	viper.SetDefault("server.rate_limit", 1000)

	viper.SetDefault("database.type", "mysql")
	viper.SetDefault("database.charset", "utf8mb4")

	if err := viper.ReadInConfig(); err != nil {
		// 用viper內部的Error defind
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
