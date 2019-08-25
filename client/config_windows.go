package client

import (
	"os"
	"path/filepath"
)

func configDir() string {
	if cfgDir != "" {
		return cfgDir
	}

	configDir := os.Getenv("APPDATA")
	if configDir == "" {
		panic("APPDATA environment variable not set - cannot locate config directory")
	}
	return filepath.Join(configDir, "timelapse")
}
