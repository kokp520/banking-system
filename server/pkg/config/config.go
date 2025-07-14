package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Logger  LoggerConfig  `mapstructure:"logger"`
	Swagger SwaggerConfig `mapstructure:"swagger"`
}

type ServerConfig struct {
	Mode         string `mapstructure:"mode"`
	Port         string `mapstructure:"port"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	RateLimit    int    `mapstructure:"rate_limit"`
}

type LoggerConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Dir    string `mapstructure:"dir"`
}

type SwaggerConfig struct {
	Path string `mapstructure:"path"`
}

func Setup(f string) (*Config, error) {
	viper.SetConfigName(f)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.read_timeout", 10)
	viper.SetDefault("server.write_timeout", 60)
	viper.SetDefault("server.rate_limit", 1000)

	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.dir", "logs")

	viper.SetDefault("swagger.path", "api/local_bank.yaml")

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
