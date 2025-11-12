package db

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ompatil-15/coconut/internal/db/model"
	"github.com/ompatil-15/coconut/internal/vault"
)

type Vault interface {
	IsUnlocked() bool
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

type EncryptedRepository struct {
	repo   Repository
	vault  Vault
	bucket string
}

func (f *RepositoryFactory) SetVault(v *vault.Vault) {
	f.vault = v
}

func NewEncryptedRepository(repo Repository, v Vault, bucket string) *EncryptedRepository {
	return &EncryptedRepository{
		repo:   repo,
		vault:  v,
		bucket: bucket,
	}
}

func (e *EncryptedRepository) Add(secret model.Secret) (string, error) {
	if !e.vault.IsUnlocked() {
		return "", fmt.Errorf("vault is locked")
	}

	data, err := json.Marshal(secret)
	if err != nil {
		return "", fmt.Errorf("marshal secret: %w", err)
	}

	enc, err := e.vault.Encrypt(string(data))
	if err != nil {
		return "", fmt.Errorf("encrypt secret: %w", err)
	}

	key := fmt.Sprintf("%v", secret.ID)
	if err := e.repo.Put(key, []byte(enc)); err != nil {
		return "", fmt.Errorf("store secret: %w", err)
	}

	return secret.ID, nil
}

func (e *EncryptedRepository) Get(key string) (*model.Secret, error) {
	if !e.vault.IsUnlocked() {
		return nil, fmt.Errorf("vault is locked")
	}

	data, err := e.repo.Get(key)
	if err != nil {
		return nil, err
	}

	dec, err := e.vault.Decrypt(string(data))
	if err != nil {
		return nil, fmt.Errorf("decrypt secret: %w", err)
	}

	var secret model.Secret
	if err := json.Unmarshal([]byte(dec), &secret); err != nil {
		return nil, fmt.Errorf("unmarshal secret: %w", err)
	}

	return &secret, nil
}

func (e *EncryptedRepository) Update(secret model.Secret) error {
	if !e.vault.IsUnlocked() {
		return fmt.Errorf("vault is locked")
	}

	secret.UpdatedAt = time.Now()
	data, err := json.Marshal(secret)
	if err != nil {
		return fmt.Errorf("marshal secret: %w", err)
	}

	enc, err := e.vault.Encrypt(string(data))
	if err != nil {
		return fmt.Errorf("encrypt secret: %w", err)
	}

	key := fmt.Sprintf("%v", secret.ID)
	return e.repo.Put(key, []byte(enc))
}

func (e *EncryptedRepository) Delete(key string) error {
	return e.repo.Delete(key)
}

func (e *EncryptedRepository) List() ([]model.Secret, error) {
	keys, err := e.repo.ListKeys()
	if err != nil {
		return nil, err
	}

	var secrets []model.Secret
	for _, k := range keys {
		secret, err := e.Get(k)
		if err != nil {
			return nil, fmt.Errorf("failed to load secret %s: %w", k, err)
		}
		secrets = append(secrets, *secret)
	}

	return secrets, nil
}
