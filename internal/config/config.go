package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	DBPath        string
	SystemBucket  string
	SecretsBucket string
	AutoLockSecs  int
	AppName       string
	Version       string
	Author        string
}

func Default() *Config {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	base := filepath.Join(home, ".coconut")

	return &Config{
		DBPath:        filepath.Join(base, "coconut.db"),
		SystemBucket:  "system",
		SecretsBucket: "secrets",
		AutoLockSecs:  300,
		AppName:       "coconut",
		Version:       "1.0.0",
		Author:        "Om Patil <patilom001@gmail.com>",
	}
}
