package session

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ompatil-15/coconut/internal/config"
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

func TestNewManager(t *testing.T) {
	repo := &mockRepository{}
	cfg := &config.Config{AutoLockSecs: 300}

	manager := NewManager(repo, cfg)

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	if manager.repo != repo {
		t.Error("Manager repository not set correctly")
	}

	if manager.cfg != cfg {
		t.Error("Manager config not set correctly")
	}
}

func TestManager_CreateSession(t *testing.T) {
	repo := &mockRepository{}
	cfg := &config.Config{AutoLockSecs: 300}
	manager := NewManager(repo, cfg)

	key := []byte("test-session-key-32-bytes-long")

	err := manager.CreateSession(key)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Verify session data was stored
	sessionData, exists := repo.data["session:data"]
	if !exists {
		t.Fatal("Session data not stored")
	}

	var session Session
	err = json.Unmarshal(sessionData, &session)
	if err != nil {
		t.Fatalf("Failed to unmarshal session data: %v", err)
	}

	if session.EncryptedKey == "" {
		t.Error("Encrypted key should not be empty")
	}

	if session.LastActivityAt.IsZero() {
		t.Error("Last activity should be set")
	}
}

func TestManager_GetCachedKey(t *testing.T) {
	repo := &mockRepository{}
	cfg := &config.Config{AutoLockSecs: 300}
	manager := NewManager(repo, cfg)

	originalKey := []byte("test-session-key-32-bytes-long")

	// Create session first
	err := manager.CreateSession(originalKey)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Get cached key
	retrievedKey, err := manager.GetCachedKey()
	if err != nil {
		t.Fatalf("GetCachedKey failed: %v", err)
	}

	if string(retrievedKey) != string(originalKey) {
		t.Errorf("Retrieved key doesn't match original key")
	}
}

func TestManager_GetCachedKey_NoSession(t *testing.T) {
	repo := &mockRepository{}
	cfg := &config.Config{AutoLockSecs: 300}
	manager := NewManager(repo, cfg)

	// Try to get cached key without creating session
	_, err := manager.GetCachedKey()
	if err == nil {
		t.Error("GetCachedKey should fail when no session exists")
	}
}

func TestManager_IsValid(t *testing.T) {
	repo := &mockRepository{}
	cfg := &config.Config{AutoLockSecs: 300}
	manager := NewManager(repo, cfg)

	// No session initially
	if manager.IsValid() {
		t.Error("Session should not be valid initially")
	}

	// Create session
	key := []byte("test-session-key-32-bytes-long")
	err := manager.CreateSession(key)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Session should be valid
	if !manager.IsValid() {
		t.Error("Session should be valid after creating")
	}
}

func TestManager_IsValid_Expired(t *testing.T) {
	repo := &mockRepository{}
	cfg := &config.Config{AutoLockSecs: 1} // 1 second timeout
	manager := NewManager(repo, cfg)

	key := []byte("test-session-key-32-bytes-long")
	err := manager.CreateSession(key)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Wait for session to expire
	time.Sleep(2 * time.Second)

	// Session should be expired
	if manager.IsValid() {
		t.Error("Session should be expired after timeout")
	}
}

func TestManager_IsValid_NoAutoLock(t *testing.T) {
	repo := &mockRepository{}
	cfg := &config.Config{AutoLockSecs: 0} // No auto-lock
	manager := NewManager(repo, cfg)

	key := []byte("test-session-key-32-bytes-long")
	err := manager.CreateSession(key)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Session should still be valid (no auto-lock)
	if !manager.IsValid() {
		t.Error("Session should remain valid when AutoLockSecs is 0")
	}
}

func TestManager_Clear(t *testing.T) {
	repo := &mockRepository{}
	cfg := &config.Config{AutoLockSecs: 300}
	manager := NewManager(repo, cfg)

	key := []byte("test-session-key-32-bytes-long")

	// Create session
	err := manager.CreateSession(key)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Verify session is valid
	if !manager.IsValid() {
		t.Error("Session should be valid before clearing")
	}

	// Clear session
	err = manager.Clear()
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	// Verify session is no longer valid
	if manager.IsValid() {
		t.Error("Session should not be valid after clearing")
	}

	// Verify session data is removed from repository
	_, exists := repo.data["session:data"]
	if exists {
		t.Error("Session data should be removed from repository")
	}
}

func TestManager_UpdateActivity(t *testing.T) {
	repo := &mockRepository{}
	cfg := &config.Config{AutoLockSecs: 300}
	manager := NewManager(repo, cfg)

	key := []byte("test-session-key-32-bytes-long")

	// Create session
	err := manager.CreateSession(key)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Get initial activity time
	sessionData1, _ := repo.data["session:data"]
	var session1 Session
	json.Unmarshal(sessionData1, &session1)
	initialActivity := session1.LastActivityAt

	// Wait a bit and update activity
	time.Sleep(10 * time.Millisecond)
	err = manager.UpdateActivity()
	if err != nil {
		t.Fatalf("UpdateActivity failed: %v", err)
	}

	// Get updated activity time
	sessionData2, _ := repo.data["session:data"]
	var session2 Session
	json.Unmarshal(sessionData2, &session2)
	updatedActivity := session2.LastActivityAt

	// Updated activity should be after initial activity
	if !updatedActivity.After(initialActivity) {
		t.Error("Updated activity time should be after initial activity time")
	}
}

func TestManager_UpdateActivity_NoSession(t *testing.T) {
	repo := &mockRepository{}
	cfg := &config.Config{AutoLockSecs: 300}
	manager := NewManager(repo, cfg)

	// Try to update activity without session
	err := manager.UpdateActivity()
	if err == nil {
		t.Error("UpdateActivity should fail when no session exists")
	}
}

func TestManager_SessionExpiration(t *testing.T) {
	repo := &mockRepository{}
	cfg := &config.Config{AutoLockSecs: 1}
	manager := NewManager(repo, cfg)

	key := []byte("test-session-key-32-bytes-long")

	// Create session
	err := manager.CreateSession(key)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Session should be valid initially
	if !manager.IsValid() {
		t.Error("Session should be valid initially")
	}

	// Wait for expiration
	time.Sleep(1500 * time.Millisecond)

	// Session should be expired
	if manager.IsValid() {
		t.Error("Session should be expired")
	}

	// Try to get cached key after expiration
	_, err = manager.GetCachedKey()
	if err == nil {
		t.Error("GetCachedKey should fail for expired session")
	}
}

func TestManager_CorruptedSessionData(t *testing.T) {
	repo := &mockRepository{
		data: map[string][]byte{
			"session:data": []byte("invalid-json"),
		},
	}
	cfg := &config.Config{AutoLockSecs: 300}
	manager := NewManager(repo, cfg)

	// IsValid should handle corrupted data gracefully
	if manager.IsValid() {
		t.Error("IsValid should return false for corrupted session data")
	}

	// GetCachedKey should fail with corrupted data
	_, err := manager.GetCachedKey()
	if err == nil {
		t.Error("GetCachedKey should fail with corrupted session data")
	}
}

func TestManager_GetRemainingTime(t *testing.T) {
	repo := &mockRepository{}
	cfg := &config.Config{AutoLockSecs: 300}
	manager := NewManager(repo, cfg)

	key := []byte("test-session-key-32-bytes-long")

	// No session initially
	remaining := manager.GetRemainingTime()
	if remaining != 0 {
		t.Error("Remaining time should be 0 when no session exists")
	}

	// Create session
	err := manager.CreateSession(key)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Should have remaining time
	remaining = manager.GetRemainingTime()
	if remaining <= 0 {
		t.Error("Should have remaining time after creating session")
	}

	if remaining > 300*time.Second {
		t.Error("Remaining time should not exceed timeout")
	}
}