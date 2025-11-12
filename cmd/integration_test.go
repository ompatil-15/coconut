// +build integration

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ompatil-15/coconut/internal/config"
	"github.com/ompatil-15/coconut/internal/db"
	"github.com/ompatil-15/coconut/internal/db/boltdb"
	"github.com/ompatil-15/coconut/internal/factory"
	"github.com/ompatil-15/coconut/internal/iostreams"
	"github.com/ompatil-15/coconut/internal/logger"
	"github.com/ompatil-15/coconut/internal/vault"
)

// Test helper to create a temporary test environment
func setupTestEnvironment(t *testing.T) (*factory.Factory, string, func()) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "coconut-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create test config
	cfg := &config.Config{
		AutoLockSecs:  300,
		DBPath:        filepath.Join(tempDir, "test.db"),
		LogPath:       filepath.Join(tempDir, "test.log"),
		SystemBucket:  "system",
		SecretsBucket: "secrets",
	}

	// Create test factory
	io := iostreams.System()
	log, err := logger.New()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	bdb, err := boltdb.NewBoltStore(cfg.DBPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	repoFactory := db.NewRepositoryFactory(bdb, nil, cfg.SystemBucket, cfg.SecretsBucket)
	systemRepo := repoFactory.NewBaseRepository(cfg.SystemBucket)

	f := &factory.Factory{
		IO:      io,
		Logger:  log,
		Config:  cfg,
		DB:      bdb,
		System:  systemRepo,
		Secrets: repoFactory.NewEncryptedRepository(cfg.SecretsBucket),
	}

	// Cleanup function
	cleanup := func() {
		if f.DB != nil {
			f.DB.Close()
		}
		if f.Logger != nil {
			f.Logger.Close()
		}
		os.RemoveAll(tempDir)
	}

	return f, tempDir, cleanup
}

// Mock stdin for password input
func mockStdin(input string) (func(), *os.File) {
	r, w, _ := os.Pipe()
	originalStdin := os.Stdin
	os.Stdin = r

	go func() {
		defer w.Close()
		w.WriteString(input)
	}()

	restore := func() {
		os.Stdin = originalStdin
		r.Close()
	}

	return restore, r
}

func TestIntegration_FullWorkflow(t *testing.T) {
	f, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Capture output
	var output bytes.Buffer
	f.IO = &iostreams.IOStreams{
		In:  strings.NewReader(""),
		Out: &output,
		Err: &output,
	}

	// Test 1: Initialize vault
	t.Run("Initialize", func(t *testing.T) {
		// Mock password input
		restore, _ := mockStdin("testpassword123\ntestpassword123\n")
		defer restore()

		err := InitializeVault(f.System, f.Logger)
		if err != nil {
			t.Fatalf("Initialize failed: %v", err)
		}

		// Verify vault was created
		if !vault.CheckVaultExists(f.System) {
			t.Error("Vault should exist after initialization")
		}
	})

	// Test 2: Try to initialize again (should fail)
	t.Run("Initialize Again", func(t *testing.T) {
		restore, _ := mockStdin("testpassword123\ntestpassword123\n")
		defer restore()

		err := InitializeVault(f.System, f.Logger)
		if err == nil {
			t.Error("Initialize should fail when vault already exists")
		}
	})

	// Test 3: Unlock vault
	t.Run("Unlock", func(t *testing.T) {
		restore, _ := mockStdin("testpassword123\n")
		defer restore()

		err := UnlockVault(f)
		if err != nil {
			t.Fatalf("Unlock failed: %v", err)
		}

		if !f.Vault.IsUnlocked() {
			t.Error("Vault should be unlocked")
		}
	})

	// Test 4: Add secrets
	secrets := []struct {
		username    string
		password    string
		url         string
		description string
	}{
		{"user1", "pass1", "https://example1.com", "Test account 1"},
		{"user2", "pass2", "https://example2.com", "Test account 2"},
		{"user3", "pass3", "", "Test account 3"},
	}

	for i, secret := range secrets {
		t.Run(fmt.Sprintf("Add Secret %d", i+1), func(t *testing.T) {
			cmd := NewAddCmd(f)
			cmd.SetArgs([]string{
				"-u", secret.username,
				"-p", secret.password,
				"-l", secret.url,
				"-d", secret.description,
			})

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("Add secret failed: %v", err)
			}
		})
	}

	// Test 5: List secrets
	t.Run("List Secrets", func(t *testing.T) {
		cmd := NewListCmd(f)
		err := cmd.Execute()
		if err != nil {
			t.Fatalf("List secrets failed: %v", err)
		}

		// Check output contains our secrets
		outputStr := output.String()
		for _, secret := range secrets {
			if !strings.Contains(outputStr, secret.username) {
				t.Errorf("Output should contain username '%s'", secret.username)
			}
		}
	})

	// Test 6: Get specific secret
	t.Run("Get Secret", func(t *testing.T) {
		cmd := NewGetCmd(f)
		cmd.SetArgs([]string{"1"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Get secret failed: %v", err)
		}
	})

	// Test 7: Update secret
	t.Run("Update Secret", func(t *testing.T) {
		cmd := NewUpdateCmd(f)
		cmd.SetArgs([]string{
			"1",
			"-u", "updated-user1",
			"-p", "updated-pass1",
		})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Update secret failed: %v", err)
		}
	})

	// Test 8: Generate password
	t.Run("Generate Password", func(t *testing.T) {
		cmd := NewGenerateCmd(f)
		cmd.SetArgs([]string{"--length", "16"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Generate password failed: %v", err)
		}
	})

	// Test 9: Lock vault
	t.Run("Lock Vault", func(t *testing.T) {
		err := LockVault(f)
		if err != nil {
			t.Fatalf("Lock vault failed: %v", err)
		}

		if f.Vault.IsUnlocked() {
			t.Error("Vault should be locked")
		}
	})

	// Test 10: Try operations when locked (should fail)
	t.Run("Operations When Locked", func(t *testing.T) {
		cmd := NewListCmd(f)
		err := cmd.Execute()
		if err == nil {
			t.Error("List should fail when vault is locked")
		}
	})

	// Test 11: Delete secret (after unlocking)
	t.Run("Delete Secret", func(t *testing.T) {
		// Unlock first
		restore, _ := mockStdin("testpassword123\n")
		defer restore()

		err := UnlockVault(f)
		if err != nil {
			t.Fatalf("Unlock failed: %v", err)
		}

		// Delete secret
		cmd := NewDeleteCmd(f)
		cmd.SetArgs([]string{"1"})

		err = cmd.Execute()
		if err != nil {
			t.Fatalf("Delete secret failed: %v", err)
		}
	})
}

func TestIntegration_WrongPassword(t *testing.T) {
	f, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Initialize vault with correct password
	restore1, _ := mockStdin("correctpassword\ncorrectpassword\n")
	defer restore1()

	err := InitializeVault(f.System, f.Logger)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Try to unlock with wrong password
	restore2, _ := mockStdin("wrongpassword\n")
	defer restore2()

	err = UnlockVault(f)
	if err == nil {
		t.Error("Unlock should fail with wrong password")
	}
}

func TestIntegration_SessionTimeout(t *testing.T) {
	f, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Set short session timeout
	f.Config.AutoLockSecs = 1

	// Initialize and unlock vault
	restore1, _ := mockStdin("testpassword\ntestpassword\n")
	defer restore1()

	err := InitializeVault(f.System, f.Logger)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	restore2, _ := mockStdin("testpassword\n")
	defer restore2()

	err = UnlockVault(f)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// Verify vault is unlocked
	if !f.Vault.IsUnlocked() {
		t.Error("Vault should be unlocked")
	}

	// Wait for session to timeout
	time.Sleep(2 * time.Second)

	// Try to add secret (should require re-authentication)
	cmd := NewAddCmd(f)
	cmd.SetArgs([]string{"-u", "test", "-p", "test"})

	err = cmd.Execute()
	if err == nil {
		t.Error("Add should fail after session timeout")
	}
}

func TestIntegration_ConfigManagement(t *testing.T) {
	f, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Initialize vault
	restore, _ := mockStdin("testpassword\ntestpassword\n")
	defer restore()

	err := InitializeVault(f.System, f.Logger)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Test config command
	t.Run("View Config", func(t *testing.T) {
		cmd := NewConfigCmd(f)
		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Config view failed: %v", err)
		}
	})

	t.Run("Set Config", func(t *testing.T) {
		cmd := NewConfigCmd(f)
		cmd.SetArgs([]string{"autoLockSecs", "600"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Config set failed: %v", err)
		}

		// Verify config was updated
		if f.Config.AutoLockSecs != 600 {
			t.Errorf("Expected AutoLockSecs 600, got %d", f.Config.AutoLockSecs)
		}
	})
}

func TestIntegration_ErrorHandling(t *testing.T) {
	f, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Test operations without initialized vault
	t.Run("Operations Without Vault", func(t *testing.T) {
		tests := []struct {
			name string
			cmd  func() error
		}{
			{"unlock", func() error { return UnlockVault(f) }},
			{"lock", func() error { return LockVault(f) }},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				err := test.cmd()
				if err == nil {
					t.Errorf("%s should fail without initialized vault", test.name)
				}
			})
		}
	})

	// Initialize vault for remaining tests
	restore, _ := mockStdin("testpassword\ntestpassword\n")
	defer restore()

	err := InitializeVault(f.System, f.Logger)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Test operations without unlocked vault
	t.Run("Operations Without Unlock", func(t *testing.T) {
		cmd := NewAddCmd(f)
		cmd.SetArgs([]string{"-u", "test", "-p", "test"})

		err := cmd.Execute()
		if err == nil {
			t.Error("Add should fail without unlocked vault")
		}
	})
}

