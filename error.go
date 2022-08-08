package toolsconfig

import (
	"errors"
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
	return "ConfigurationError: " + e.Err.Error() + missing + " - check config file"
}

func wrapErr(err error, missing ...string) *ConfigError {
	return &ConfigError{Err: err, Missing: missing}
}

var errNotFound = errors.New("not found")
