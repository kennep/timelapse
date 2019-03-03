package client

import (
	"os"
	"path/filepath"
)

func configDir() string {
	if cfgDir != "" {
		return cfgDir
	}

	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home := os.Getenv("HOME")
		if home == "" {
			panic("HOME environment variable not set - cannot locate config directory")
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "timelapse")
}
