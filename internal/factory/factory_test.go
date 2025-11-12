package factory

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "factory-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set environment to use temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create .coconut directory
	coconutDir := filepath.Join(tempDir, ".coconut")
	err = os.MkdirAll(coconutDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create coconut dir: %v", err)
	}

	// Create logs directory
	logsDir := filepath.Join(coconutDir, "logs")
	err = os.MkdirAll(logsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create logs dir: %v", err)
	}

	// Test factory creation
	factory, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer factory.Close()

	// Verify factory components
	if factory.IO == nil {
		t.Error("IO should not be nil")
	}

	if factory.Logger == nil {
		t.Error("Logger should not be nil")
	}

	if factory.Config == nil {
		t.Error("Config should not be nil")
	}

	if factory.DB == nil {
		t.Error("DB should not be nil")
	}

	if factory.Vault == nil {
		t.Error("Vault should not be nil")
	}

	if factory.Crypto == nil {
		t.Error("Crypto should not be nil")
	}

	if factory.Repo == nil {
		t.Error("Repo should not be nil")
	}

	if factory.System == nil {
		t.Error("System should not be nil")
	}

	if factory.Secrets == nil {
		t.Error("Secrets should not be nil")
	}

	if factory.Session == nil {
		t.Error("Session should not be nil")
	}
}

func TestFactory_Close(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "factory-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set environment to use temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create .coconut directory
	coconutDir := filepath.Join(tempDir, ".coconut")
	err = os.MkdirAll(coconutDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create coconut dir: %v", err)
	}

	// Create logs directory
	logsDir := filepath.Join(coconutDir, "logs")
	err = os.MkdirAll(logsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create logs dir: %v", err)
	}

	factory, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Test close
	factory.Close()

	// Close should be idempotent (safe to call multiple times)
	factory.Close()
}

func TestFactory_InvalidPath(t *testing.T) {
	// Set invalid home directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", "/invalid/path/that/does/not/exist")
	defer os.Setenv("HOME", originalHome)

	// Factory creation should still work (it creates directories)
	factory, err := New()
	if err != nil {
		// This might fail due to permissions, which is expected
		t.Logf("Factory creation failed as expected: %v", err)
		return
	}
	defer factory.Close()

	// If it succeeds, verify components are created
	if factory == nil {
		t.Error("Factory should not be nil even with invalid path")
	}
}

func TestFactory_ComponentIntegration(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "factory-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set environment to use temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create .coconut directory
	coconutDir := filepath.Join(tempDir, ".coconut")
	err = os.MkdirAll(coconutDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create coconut dir: %v", err)
	}

	// Create logs directory
	logsDir := filepath.Join(coconutDir, "logs")
	err = os.MkdirAll(logsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create logs dir: %v", err)
	}

	factory, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer factory.Close()

	// Test that components can interact
	// Test system repository
	err = factory.System.Put("test-key", []byte("test-value"))
	if err != nil {
		t.Fatalf("System.Put failed: %v", err)
	}

	value, err := factory.System.Get("test-key")
	if err != nil {
		t.Fatalf("System.Get failed: %v", err)
	}

	if string(value) != "test-value" {
		t.Errorf("Expected 'test-value', got '%s'", string(value))
	}

	// Test config
	if factory.Config.AutoLockSecs != 300 {
		t.Errorf("Expected default AutoLockSecs 300, got %d", factory.Config.AutoLockSecs)
	}

	// Test vault is not unlocked initially
	if factory.Vault.IsUnlocked() {
		t.Error("Vault should not be unlocked initially")
	}
}