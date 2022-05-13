package config

import (
	"path/filepath"

	"github.com/crochee/devt/pkg/config"
)

// LoadConfig init Config
func LoadConfig(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	return config.LoadConfig(
		config.WithConfigFile(absPath),
	)
}
