// internal/config/config.go
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Environment string `mapstructure:"environment"`

	Server struct {
		Port     string `mapstructure:"port"`
		Host     string `mapstructure:"host"`
		BasePath string `mapstructure:"base_path"`
	} `mapstructure:"server"`

	Database struct {
		Host     string `mapstructure:"host"`
		Port     string `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Name     string `mapstructure:"name"`
		SSLMode  string `mapstructure:"sslmode"`
	} `mapstructure:"database"`

	JWT struct {
		Secret     string `mapstructure:"secret"`
		ExpireHour int    `mapstructure:"expire_hour"`
	} `mapstructure:"jwt"`

	MarketData struct {
		Provider  string `mapstructure:"provider"`
		APIKey    string `mapstructure:"api_key"`
		APISecret string `mapstructure:"api_secret"`
		WSPort    string `mapstructure:"ws_port"`
	} `mapstructure:"market_data"`

	Broker struct {
		Provider  string `mapstructure:"provider"`
		APIKey    string `mapstructure:"api_key"`
		APISecret string `mapstructure:"api_secret"`
		IsPaper   bool   `mapstructure:"is_paper"`
	} `mapstructure:"broker"`
}

func Load() (*Config, error) {
	viper.SetConfigName("app")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath("../../configs")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("SENTINEL")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Load environment-specific config
	env := viper.GetString("environment")
	if env != "" {
		viper.SetConfigName(fmt.Sprintf("app.%s", env))
		if err := viper.MergeInConfig(); err != nil {
			// It's okay if the environment config doesn't exist
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("failed to read environment config: %w", err)
			}
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
