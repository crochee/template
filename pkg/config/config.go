package config

import (
	"context"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

type option struct {
	cfg        string
	name       string
	envPrefix  string
	configType string
}

type Option func(*option)

func WithConfigFile(cfg string) Option {
	return func(o *option) {
		o.cfg = cfg
	}
}

func WithConfigType(configType string) Option {
	return func(o *option) {
		o.configType = configType
	}
}

func WithName(name string) Option {
	return func(o *option) {
		o.name = name
	}
}

func WithEnvPrefix(envPrefix string) Option {
	return func(o *option) {
		o.envPrefix = envPrefix
	}
}

// LoadConfig init Config
func LoadConfig(opts ...Option) error {
	o := &option{
		envPrefix:  "cloud",
		configType: "yaml",
	}

	for _, opt := range opts {
		opt(o)
	}
	if o.cfg != "" {
		// Use config file from the flag.
		viper.SetConfigFile(o.cfg)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			return err
		}
		// Search config in home directory with name ".migrate" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(o.name)
		viper.SetConfigType(o.configType)
	}

	viper.SetEnvPrefix(o.envPrefix) // set environment variables prefix to avoid conflict
	viper.AutomaticEnv()            // read in environment variables that match

	// If a config file is found, read it in.
	return viper.ReadInConfig()
}

type SyncHandler interface {
	Sync(context.Context) error
}

type Getter interface {
	Get(ctx context.Context, selectRemote bool) error
}

// Sync 同步所有处理者
func Sync(ctx context.Context, handlers ...SyncHandler) error {
	for _, handler := range handlers {
		if err := handler.Sync(ctx); err != nil {
			return err
		}
	}
	return nil
}

func Get(ctx context.Context, selectRemote bool, getters ...Getter) error {
	for _, getter := range getters {
		if err := getter.Get(ctx, selectRemote); err != nil {
			return err
		}
	}
	return nil
}
