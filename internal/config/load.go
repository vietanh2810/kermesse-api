package config

import (
	"fmt"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func Load(configFile string) (*AppConfig, error) {
	viper.SetConfigFile(configFile)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("viper.ReadInConfig -> %w", err)
	}

	var conf *AppConfig
	if err := viper.Unmarshal(&conf); err != nil {
		return nil, fmt.Errorf("viper.Unmarshal -> %w", err)
	}

	if err := conf.validateConfig(); err != nil {
		return nil, fmt.Errorf("conf.validateConfig -> %w", err)
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		zap.L().Info(
			"config file changed",
			zap.String("fileName", e.Name),
			zap.Any("operation", e.Op),
		)
	})

	return conf, nil
}
