package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// Mock repository for testing
type mockRepository struct {
	data map[string][]byte
}

func (m *mockRepository) Put(key string, value []byte) error {
	if m.data == nil {
		m.data = make(map[string][]byte)
	}
	m.data[key] = value
	return nil
}

func (m *mockRepository) Get(key string) ([]byte, error) {
	if data, exists := m.data[key]; exists {
		return data, nil
	}
	return nil, errors.New("key not found")
}

func (m *mockRepository) Delete(key string) error {
	if m.data == nil {
		return errors.New("key not found")
	}
	delete(m.data, key)
	return nil
}

func (m *mockRepository) ListKeys() ([]string, error) {
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys, nil
}

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg == nil {
		t.Fatal("Default() returned nil")
	}

	// Test default values
	if cfg.AutoLockSecs != 300 {
		t.Errorf("Expected AutoLockSecs 300, got %d", cfg.AutoLockSecs)
	}

	// Test default paths
	homeDir, _ := os.UserHomeDir()
	expectedDBPath := filepath.Join(homeDir, ".coconut", "coconut.db")
	if cfg.DBPath != expectedDBPath {
		t.Errorf("Expected DBPath '%s', got '%s'", expectedDBPath, cfg.DBPath)
	}

	// Test bucket names
	if cfg.SystemBucket != "system" {
		t.Errorf("Expected SystemBucket 'system', got '%s'", cfg.SystemBucket)
	}

	if cfg.SecretsBucket != "secrets" {
		t.Errorf("Expected SecretsBucket 'secrets', got '%s'", cfg.SecretsBucket)
	}
}

func TestSave(t *testing.T) {
	repo := &mockRepository{}
	cfg := &Config{
		AutoLockSecs:  600,
		DBPath:        "/test/path/db",
		SystemBucket:  "test-system",
		SecretsBucket: "test-secrets",
	}

	err := Save(repo, cfg)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify data was saved
	data, exists := repo.data["config:data"]
	if !exists {
		t.Fatal("Config data not saved to repository")
	}

	// Verify JSON structure
	var savedConfig Config
	err = json.Unmarshal(data, &savedConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal saved config: %v", err)
	}

	if savedConfig.AutoLockSecs != cfg.AutoLockSecs {
		t.Errorf("Expected AutoLockSecs %d, got %d", cfg.AutoLockSecs, savedConfig.AutoLockSecs)
	}
}

func TestLoad(t *testing.T) {
	// Test loading existing config
	existingConfig := &Config{
		AutoLockSecs:  900,
		DBPath:        "/custom/path/db",
		SystemBucket:  "custom-system",
		SecretsBucket: "custom-secrets",
	}

	configData, _ := json.Marshal(existingConfig)
	repo := &mockRepository{
		data: map[string][]byte{
			"config:data": configData,
		},
	}

	loadedConfig, err := Load(repo)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loadedConfig.AutoLockSecs != existingConfig.AutoLockSecs {
		t.Errorf("Expected AutoLockSecs %d, got %d", existingConfig.AutoLockSecs, loadedConfig.AutoLockSecs)
	}

	if loadedConfig.DBPath != existingConfig.DBPath {
		t.Errorf("Expected DBPath '%s', got '%s'", existingConfig.DBPath, loadedConfig.DBPath)
	}
}

func TestLoad_NotFound(t *testing.T) {
	// Test loading when config doesn't exist (should return default)
	repo := &mockRepository{}

	loadedConfig, err := Load(repo)
	if err != nil {
		t.Fatalf("Load should not fail when config doesn't exist: %v", err)
	}

	defaultConfig := Default()
	if loadedConfig.AutoLockSecs != defaultConfig.AutoLockSecs {
		t.Errorf("Expected default AutoLockSecs %d, got %d", defaultConfig.AutoLockSecs, loadedConfig.AutoLockSecs)
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	// Test loading with invalid JSON
	repo := &mockRepository{
		data: map[string][]byte{
			"config:data": []byte("invalid-json"),
		},
	}

	_, err := Load(repo)
	if err == nil {
		t.Error("Load should fail with invalid JSON")
	}
}

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		valid  bool
	}{
		{
			name: "valid config",
			config: &Config{
				AutoLockSecs:  300,
				DBPath:        "/valid/path",
				SystemBucket:  "system",
				SecretsBucket: "secrets",
			},
			valid: true,
		},
		{
			name: "zero autolock",
			config: &Config{
				AutoLockSecs:  0,
				DBPath:        "/valid/path",
				SystemBucket:  "system",
				SecretsBucket: "secrets",
			},
			valid: true, // Zero is valid (no auto-lock)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{}
			err := Save(repo, tt.config)

			if tt.valid && err != nil {
				t.Errorf("Save should succeed for valid config: %v", err)
			}
			if !tt.valid && err == nil {
				t.Error("Save should fail for invalid config")
			}
		})
	}
}

func TestConfig_RoundTrip(t *testing.T) {
	// Test save and load round trip
	repo := &mockRepository{}
	originalConfig := &Config{
		AutoLockSecs:  1800,
		DBPath:        "/test/roundtrip/db",
		SystemBucket:  "rt-system",
		SecretsBucket: "rt-secrets",
	}

	// Save config
	err := Save(repo, originalConfig)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load config
	loadedConfig, err := Load(repo)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Compare all fields
	if loadedConfig.AutoLockSecs != originalConfig.AutoLockSecs {
		t.Errorf("AutoLockSecs mismatch: expected %d, got %d", originalConfig.AutoLockSecs, loadedConfig.AutoLockSecs)
	}

	if loadedConfig.DBPath != originalConfig.DBPath {
		t.Errorf("DBPath mismatch: expected '%s', got '%s'", originalConfig.DBPath, loadedConfig.DBPath)
	}

	if loadedConfig.SystemBucket != originalConfig.SystemBucket {
		t.Errorf("SystemBucket mismatch: expected '%s', got '%s'", originalConfig.SystemBucket, loadedConfig.SystemBucket)
	}

	if loadedConfig.SecretsBucket != originalConfig.SecretsBucket {
		t.Errorf("SecretsBucket mismatch: expected '%s', got '%s'", originalConfig.SecretsBucket, loadedConfig.SecretsBucket)
	}
}

func TestConfig_DefaultPaths(t *testing.T) {
	cfg := Default()

	// Verify paths contain expected components
	if !filepath.IsAbs(cfg.DBPath) {
		t.Error("DBPath should be absolute")
	}

	// Verify paths contain .coconut directory
	if !contains(cfg.DBPath, ".coconut") {
		t.Error("DBPath should contain .coconut directory")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[:len(substr)+1] == substr+"/" || s[len(s)-len(substr)-1:] == "/"+substr || contains(s[1:], substr))))
}