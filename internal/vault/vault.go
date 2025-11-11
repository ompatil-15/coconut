package vault

import (
	"errors"

	"github.com/ompatil-15/coconut/internal/crypto"
)

const (
	saltKey              = "salt"
	verificationTokenKey = "vault_verification"
)

// Verification token is a constant that we encrypt to verify password correctness
const verificationTokenValue = "coconut-vault-v1-verification"

type Vault struct {
	strategy crypto.CryptoStrategy
	key      []byte
	salt     []byte
	unlocked bool
}

type SystemReader interface {
	Get(key string) ([]byte, error)
}

type SaltStore interface {
	Get(key string) ([]byte, error)
	Put(key string, value []byte) error
}

func NewVault(strategy crypto.CryptoStrategy, salt []byte) *Vault {
	return &Vault{
		strategy: strategy,
		salt:     salt,
		unlocked: false,
	}
}

func LoadVaultConfig(systemRepo SystemReader) ([]byte, error) {
	salt, _ := systemRepo.Get(saltKey)
	return salt, nil
}

func (v *Vault) Unlock(derivedKey []byte) {
	v.key = derivedKey
	v.unlocked = true
}

func (v *Vault) Lock() {
	v.unlocked = false
	if v.key != nil {
		for i := range v.key {
			v.key[i] = 0
		}
	}
	v.key = nil
}

func (v *Vault) IsUnlocked() bool {
	return v.unlocked
}

func (v *Vault) Encrypt(plaintext string) (string, error) {
	if !v.unlocked {
		return "", errors.New("vault locked")
	}
	return v.strategy.Encrypt(v.key, plaintext)
}

func (v *Vault) Decrypt(ciphertext string) (string, error) {
	if !v.unlocked {
		return "", errors.New("vault locked")
	}
	return v.strategy.Decrypt(v.key, ciphertext)
}

// CreateVerificationToken creates and encrypts a verification token for password validation.
// This should be called during vault initialization.
// Returns the encrypted token to be stored in the database.
func (v *Vault) CreateVerificationToken() (string, error) {
	if !v.unlocked {
		return "", errors.New("vault must be unlocked to create verification token")
	}
	return v.Encrypt(verificationTokenValue)
}

// VerifyPassword verifies that the vault was unlocked with the correct password
// by attempting to decrypt and validate the stored verification token.
// Returns nil if password is correct, error otherwise.
func (v *Vault) VerifyPassword(encryptedToken string) error {
	if !v.unlocked {
		return errors.New("vault must be unlocked to verify password")
	}

	decrypted, err := v.Decrypt(encryptedToken)
	if err != nil {
		return errors.New("incorrect master password")
	}

	if decrypted != verificationTokenValue {
		return errors.New("vault verification failed - possible corruption")
	}

	return nil
}

var ErrVaultNotFound = errors.New("vault not initialized")

// CheckVaultExists checks if a vault has been initialized
// Returns true if vault exists, false otherwise
func CheckVaultExists(systemRepo SystemReader) bool {
	salt, _ := systemRepo.Get(saltKey)
	verificationTokenKey, _ := systemRepo.Get(verificationTokenKey)
	return len(salt) > 0 && len(verificationTokenKey) > 0
}

// UnlockWithKey creates a new vault and unlocks it with the provided key
// This is the core vault unlocking operation, independent of how the key was obtained
func UnlockWithKey(strategy crypto.CryptoStrategy, salt []byte, key []byte) *Vault {
	v := NewVault(strategy, salt)
	v.Unlock(key)
	return v
}

// VerifyVaultPassword verifies that a key correctly unlocks the vault
// by attempting to decrypt the verification token
func VerifyVaultPassword(systemRepo SystemReader, vault *Vault) error {
	encryptedToken, err := systemRepo.Get(verificationTokenKey)
	if err != nil {
		return nil
	}

	return vault.VerifyPassword(string(encryptedToken))
}
