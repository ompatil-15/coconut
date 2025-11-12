package config

import (
	"encoding/json"

	"github.com/ompatil-15/coconut/internal/db"
)

const configDataKey = "config:data"

type storedConfig struct {
	AutoLockSecs  int    `json:"autoLockSecs"`
	DBPath        string `json:"dbPath"`
	SystemBucket  string `json:"systemBucket"`
	SecretsBucket string `json:"secretsBucket"`
}

// Load retrieves configuration from the system repository, applying defaults when not present.
func Load(systemRepo db.Repository) (*Config, error) {
	cfg := Default()

	data, err := systemRepo.Get(configDataKey)
	if err != nil || len(data) == 0 {
		return cfg, nil
	}

	var stored storedConfig
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, err
	}

	cfg.AutoLockSecs = stored.AutoLockSecs
	if stored.DBPath != "" {
		cfg.DBPath = stored.DBPath
	}
	if stored.SystemBucket != "" {
		cfg.SystemBucket = stored.SystemBucket
	}
	if stored.SecretsBucket != "" {
		cfg.SecretsBucket = stored.SecretsBucket
	}

	return cfg, nil
}

// Save persists configuration values that can change at runtime.
func Save(systemRepo db.Repository, cfg *Config) error {
	stored := storedConfig{
		AutoLockSecs:  cfg.AutoLockSecs,
		DBPath:        cfg.DBPath,
		SystemBucket:  cfg.SystemBucket,
		SecretsBucket: cfg.SecretsBucket,
	}

	payload, err := json.Marshal(stored)
	if err != nil {
		return err
	}

	return systemRepo.Put(configDataKey, payload)
}
