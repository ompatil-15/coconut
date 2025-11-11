package session

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ompatil-15/coconut/internal/config"
	"github.com/ompatil-15/coconut/internal/crypto"
	"github.com/ompatil-15/coconut/internal/db"
)

// Session represents an authenticated vault session with cached credentials.
// The session expires after TimeoutSeconds of inactivity (no commands executed).
// Each command execution updates LastActivityAt, extending the session.
type Session struct {
	UnlockedAt     time.Time `json:"unlocked_at"`      // When the vault was first unlocked
	LastActivityAt time.Time `json:"last_activity_at"` // Last command execution time (used for inactivity timeout)
	TimeoutSeconds int       `json:"timeout_seconds"`  // Inactivity timeout in seconds
	EncryptedKey   string    `json:"encrypted_key"`    // Vault key encrypted with session key (nonce embedded)
}

const (
	sessionDataKey = "session:data"
	sessionKeyKey  = "session:key"
)

type Manager struct {
	repo db.Repository
	cfg  *config.Config
}

func NewManager(repo db.Repository, cfg *config.Config) *Manager {
	return &Manager{
		repo: repo,
		cfg:  cfg,
	}
}

func (m *Manager) CreateSession(vaultKey []byte) error {
	sessionKey := make([]byte, 32)
	if _, err := rand.Read(sessionKey); err != nil {
		return fmt.Errorf("failed to generate session key: %w", err)
	}

	aesGCM := crypto.NewAESGCM()
	encryptedKey, err := aesGCM.Encrypt(sessionKey, base64.StdEncoding.EncodeToString(vaultKey))
	if err != nil {
		return fmt.Errorf("failed to encrypt vault key: %w", err)
	}

	now := time.Now()
	session := Session{
		UnlockedAt:     now,
		LastActivityAt: now,
		TimeoutSeconds: m.cfg.AutoLockSecs,
		EncryptedKey:   encryptedKey,
	}

	if err := m.saveSession(&session); err != nil {
		return err
	}

	if err := m.saveSessionKey(sessionKey); err != nil {
		return err
	}

	return nil
}

// IsValid checks if the current session is still valid (not expired).
// A session is valid if the time since LastActivityAt is less than the timeout.
func (m *Manager) IsValid() bool {
	session, err := m.loadSession()
	if err != nil {
		return false
	}

	timeoutSeconds := session.TimeoutSeconds
	if m.cfg.AutoLockSecs > 0 && m.cfg.AutoLockSecs < timeoutSeconds {
		timeoutSeconds = m.cfg.AutoLockSecs
	}

	elapsed := time.Since(session.LastActivityAt)
	timeout := time.Duration(timeoutSeconds) * time.Second

	return elapsed < timeout
}

// UpdateActivity updates the last activity timestamp to now.
// This should be called on every command execution to track user activity.
// Extends the session timeout by resetting the inactivity timer.
func (m *Manager) UpdateActivity() error {
	session, err := m.loadSession()
	if err != nil {
		return fmt.Errorf("no active session to update: %w", err)
	}

	session.LastActivityAt = time.Now()
	return m.saveSession(session)
}

// GetCachedKey retrieves the vault key from the session cache
// Returns nil if session is invalid or expired
func (m *Manager) GetCachedKey() ([]byte, error) {
	if !m.IsValid() {
		return nil, fmt.Errorf("session expired or invalid")
	}

	session, err := m.loadSession()
	if err != nil {
		return nil, err
	}

	sessionKey, err := m.loadSessionKey()
	if err != nil {
		return nil, fmt.Errorf("failed to load session key: %w", err)
	}

	aesGCM := crypto.NewAESGCM()
	decryptedStr, err := aesGCM.Decrypt(sessionKey, session.EncryptedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt vault key: %w", err)
	}

	vaultKey, err := base64.StdEncoding.DecodeString(decryptedStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode vault key: %w", err)
	}

	return vaultKey, nil
}

// Clear removes the session data (explicit lock)
func (m *Manager) Clear() error {
	_ = m.repo.Delete(sessionDataKey)
	_ = m.repo.Delete(sessionKeyKey)
	return nil
}

// GetRemainingTime returns the time remaining before session expires due to inactivity.
// Calculated as: timeout - (now - LastActivityAt)
func (m *Manager) GetRemainingTime() time.Duration {
	session, err := m.loadSession()
	if err != nil {
		return 0
	}

	timeoutSeconds := session.TimeoutSeconds
	if m.cfg.AutoLockSecs > 0 && m.cfg.AutoLockSecs < timeoutSeconds {
		timeoutSeconds = m.cfg.AutoLockSecs
	}

	elapsed := time.Since(session.LastActivityAt)
	timeout := time.Duration(timeoutSeconds) * time.Second
	remaining := timeout - elapsed

	if remaining < 0 {
		return 0
	}
	return remaining
}

// saveSession persists the session to the repository
func (m *Manager) saveSession(session *Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	return m.repo.Put(sessionDataKey, data)
}

// loadSession reads the session from the repository
func (m *Manager) loadSession() (*Session, error) {
	data, err := m.repo.Get(sessionDataKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read session: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// saveSessionKey saves the session key to the repository
func (m *Manager) saveSessionKey(key []byte) error {
	return m.repo.Put(sessionKeyKey, key)
}

// loadSessionKey reads the session key from the repository
func (m *Manager) loadSessionKey() ([]byte, error) {
	data, err := m.repo.Get(sessionKeyKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read session key: %w", err)
	}
	return data, nil
}


