//go:build !windows
// +build !windows

package toolsconfig

import (
	"fmt"
	"os"
)

func checkConfigFilePermissions(file *string) error {
	if file == nil {
		return fmt.Errorf("config file is required")
	}
	info, err := os.Stat(*file)

	if err != nil {
		return wrapErr(fmt.Errorf("could not stat configuration file %q: %w", *file, err))
	}
	permissions := info.Mode().Perm()
	if permissions != 0o600 {
		return fmt.Errorf("incorrect permissions %s (0%o), must be 0600 for '%s'", permissions, permissions, *file)
	}
	return nil
}
