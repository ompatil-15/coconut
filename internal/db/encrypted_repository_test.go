package db

import (
	"errors"
	"testing"
	"time"

	"github.com/ompatil-15/coconut/internal/db/model"
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
	if _, exists := m.data[key]; !exists {
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

// Mock vault for testing
type mockVault struct {
	unlocked    bool
	encryptFunc func(string) (string, error)
	decryptFunc func(string) (string, error)
}

// Ensure mockVault implements Vault
var _ Vault = (*mockVault)(nil)

func (m *mockVault) IsUnlocked() bool {
	return m.unlocked
}

func (m *mockVault) Encrypt(plaintext string) (string, error) {
	if !m.unlocked {
		return "", errors.New("vault locked")
	}
	if m.encryptFunc != nil {
		return m.encryptFunc(plaintext)
	}
	return "encrypted:" + plaintext, nil
}

func (m *mockVault) Decrypt(ciphertext string) (string, error) {
	if !m.unlocked {
		return "", errors.New("vault locked")
	}
	if m.decryptFunc != nil {
		return m.decryptFunc(ciphertext)
	}
	if len(ciphertext) > 10 && ciphertext[:10] == "encrypted:" {
		return ciphertext[10:], nil
	}
	return "", errors.New("invalid ciphertext")
}

func TestEncryptedRepository_Add(t *testing.T) {
	baseRepo := &mockRepository{}
	vault := &mockVault{unlocked: true}
	repo := NewEncryptedRepository(baseRepo, vault, "test-bucket")

	secret := model.Secret{
		ID:          "test-id",
		Username:    "testuser",
		Password:    "testpass",
		URL:         "https://example.com",
		Description: "test description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Test successful add
	key, err := repo.Add(secret)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if key == "" {
		t.Error("Add should return non-empty key")
	}

	// Verify data was encrypted and stored
	storedData, exists := baseRepo.data[key]
	if !exists {
		t.Error("Data should be stored in base repository")
	}

	if string(storedData)[:10] != "encrypted:" {
		t.Error("Data should be encrypted")
	}

	// Test add when vault is locked
	vault.unlocked = false
	_, err = repo.Add(secret)
	if err == nil {
		t.Error("Add should fail when vault is locked")
	}
}

func TestEncryptedRepository_Get(t *testing.T) {
	baseRepo := &mockRepository{}
	vault := &mockVault{unlocked: true}
	repo := NewEncryptedRepository(baseRepo, vault, "test-bucket")

	// Create test secret
	secret := model.Secret{
		ID:       "test-id",
		Username: "testuser",
		Password: "testpass",
	}

	// Add secret first
	key, err := repo.Add(secret)
	if err != nil {
		t.Fatalf("Failed to add secret: %v", err)
	}

	// Test successful get
	retrieved, err := repo.Get(key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.ID != secret.ID {
		t.Errorf("Expected ID '%s', got '%s'", secret.ID, retrieved.ID)
	}
	if retrieved.Username != secret.Username {
		t.Errorf("Expected username '%s', got '%s'", secret.Username, retrieved.Username)
	}

	// Test get with non-existent key
	_, err = repo.Get("non-existent")
	if err == nil {
		t.Error("Get should fail with non-existent key")
	}

	// Test get when vault is locked
	vault.unlocked = false
	_, err = repo.Get(key)
	if err == nil {
		t.Error("Get should fail when vault is locked")
	}
}

func TestEncryptedRepository_Update(t *testing.T) {
	baseRepo := &mockRepository{}
	vault := &mockVault{unlocked: true}
	repo := NewEncryptedRepository(baseRepo, vault, "test-bucket")

	// Create and add initial secret
	secret := model.Secret{
		ID:       "test-id",
		Username: "testuser",
		Password: "testpass",
	}

	key, err := repo.Add(secret)
	if err != nil {
		t.Fatalf("Failed to add secret: %v", err)
	}

	// Update secret
	secret.Username = "updateduser"
	secret.Password = "updatedpass"
	secret.UpdatedAt = time.Now()

	err = repo.Update(secret)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	retrieved, err := repo.Get(key)
	if err != nil {
		t.Fatalf("Failed to get updated secret: %v", err)
	}

	if retrieved.Username != "updateduser" {
		t.Errorf("Expected updated username 'updateduser', got '%s'", retrieved.Username)
	}

	// Test update when vault is locked
	vault.unlocked = false
	err = repo.Update(secret)
	if err == nil {
		t.Error("Update should fail when vault is locked")
	}
}

func TestEncryptedRepository_Delete(t *testing.T) {
	baseRepo := &mockRepository{}
	vault := &mockVault{unlocked: true}
	repo := NewEncryptedRepository(baseRepo, vault, "test-bucket")

	// Create and add secret
	secret := model.Secret{
		ID:       "test-id",
		Username: "testuser",
		Password: "testpass",
	}

	key, err := repo.Add(secret)
	if err != nil {
		t.Fatalf("Failed to add secret: %v", err)
	}

	// Verify secret exists
	_, err = repo.Get(key)
	if err != nil {
		t.Fatalf("Secret should exist before delete: %v", err)
	}

	// Delete secret
	err = repo.Delete(key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify secret is deleted
	_, err = repo.Get(key)
	if err == nil {
		t.Error("Secret should not exist after delete")
	}

	// Test delete non-existent key
	err = repo.Delete("non-existent")
	if err == nil {
		t.Error("Delete should fail with non-existent key")
	}
}

func TestEncryptedRepository_List(t *testing.T) {
	baseRepo := &mockRepository{}
	vault := &mockVault{unlocked: true}
	repo := NewEncryptedRepository(baseRepo, vault, "test-bucket")

	// Add multiple secrets
	secrets := []model.Secret{
		{ID: "1", Username: "user1", Password: "pass1"},
		{ID: "2", Username: "user2", Password: "pass2"},
		{ID: "3", Username: "user3", Password: "pass3"},
	}

	for _, secret := range secrets {
		_, err := repo.Add(secret)
		if err != nil {
			t.Fatalf("Failed to add secret: %v", err)
		}
	}

	// List all secrets
	listed, err := repo.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(listed) != len(secrets) {
		t.Errorf("Expected %d secrets, got %d", len(secrets), len(listed))
	}

	// Verify all secrets are present
	foundIDs := make(map[string]bool)
	for _, secret := range listed {
		foundIDs[secret.ID] = true
	}

	for _, original := range secrets {
		if !foundIDs[original.ID] {
			t.Errorf("Secret with ID '%s' not found in list", original.ID)
		}
	}

	// Test list when vault is locked
	vault.unlocked = false
	_, err = repo.List()
	if err == nil {
		t.Error("List should fail when vault is locked")
	}
}

func TestEncryptedRepository_EncryptionFailure(t *testing.T) {
	baseRepo := &mockRepository{}
	vault := &mockVault{
		unlocked: true,
		encryptFunc: func(s string) (string, error) {
			return "", errors.New("encryption failed")
		},
	}
	repo := NewEncryptedRepository(baseRepo, vault, "test-bucket")

	secret := model.Secret{
		ID:       "test-id",
		Username: "testuser",
		Password: "testpass",
	}

	// Test add with encryption failure
	_, err := repo.Add(secret)
	if err == nil {
		t.Error("Add should fail when encryption fails")
	}
}

func TestEncryptedRepository_DecryptionFailure(t *testing.T) {
	baseRepo := &mockRepository{}
	vault := &mockVault{unlocked: true}
	repo := NewEncryptedRepository(baseRepo, vault, "test-bucket")

	// Add secret successfully
	secret := model.Secret{
		ID:       "test-id",
		Username: "testuser",
		Password: "testpass",
	}

	key, err := repo.Add(secret)
	if err != nil {
		t.Fatalf("Failed to add secret: %v", err)
	}

	// Change vault to fail decryption
	vault.decryptFunc = func(s string) (string, error) {
		return "", errors.New("decryption failed")
	}

	// Test get with decryption failure
	_, err = repo.Get(key)
	if err == nil {
		t.Error("Get should fail when decryption fails")
	}
}

func TestEncryptedRepository_InvalidJSON(t *testing.T) {
	baseRepo := &mockRepository{
		data: map[string][]byte{
			"test-key": []byte("encrypted:invalid-json"),
		},
	}
	vault := &mockVault{unlocked: true}
	repo := NewEncryptedRepository(baseRepo, vault, "test-bucket")

	// Test get with invalid JSON
	_, err := repo.Get("test-key")
	if err == nil {
		t.Error("Get should fail with invalid JSON")
	}
}
