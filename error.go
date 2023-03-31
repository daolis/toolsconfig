package toolsconfig

import (
	"errors"
	"path/filepath"
	"strings"
)

type ConfigError struct {
	Missing []string
	Err     error
}

func (e ConfigError) Error() string {
	var missing string
	if len(e.Missing) > 0 {
		missing = " [" + strings.Join(e.Missing, ", ") + "]"
	}
	var configFilePath string
	if configDirectory != nil && configFile != nil {
		name, err := configFileName(*configDirectory, *configFile)
		if err == nil {
			absConfigFilePath, err := filepath.Abs(*name)
			if err != nil {
				panic(err)
			}
			configFilePath = absConfigFilePath
		}
	}
	return "ConfigurationError: " + e.Err.Error() + missing + " - check config file '" + configFilePath + "'"
}

func wrapErr(err error, missing ...string) *ConfigError {
	return &ConfigError{Err: err, Missing: missing}
}

var errNotFound = errors.New("not found")
