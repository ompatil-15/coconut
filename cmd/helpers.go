package cmd

import (
	"fmt"
	"os"

	"github.com/ompatil-15/coconut/internal/crypto"
	"github.com/ompatil-15/coconut/internal/factory"
	"github.com/ompatil-15/coconut/internal/vault"
	"golang.org/x/term"
)

// EnsureVaultUnlocked orchestrates vault unlocking with session management.
// This is a command-layer function that coordinates between vault, session, and factory.
//
// Flow:
//  1. Check vault exists (vault package)
//  2. Try cached session key (session package)
//  3. If no session, prompt for password (command layer)
//  4. Unlock vault (vault package)
//  5. Update factory state (command layer)
func EnsureVaultUnlocked(f *factory.Factory) error {
	// Check if vault exists (vault package responsibility)
	if !vault.CheckVaultExists(f.System) {
		fmt.Println("Error: No vault found")
		fmt.Println("")
		fmt.Println("To create a new vault, run:")
		fmt.Println("  coconut init")
		fmt.Println("")
		return vault.ErrVaultNotFound
	}

	// Get vault salt
	salt, err := f.System.Get("salt")
	if err != nil {
		return fmt.Errorf("failed to retrieve vault salt: %w", err)
	}

	var vaultKey []byte
	var createSession bool

	// Try to get cached key from valid session
	if cachedKey, err := f.Session.GetCachedKey(); err == nil {
		// Session is valid - use cached key
		vaultKey = cachedKey
		createSession = false

		// Update session activity timestamp
		f.Session.UpdateActivity()
	} else {
		// No valid session - prompt for password and derive key
		promptedKey, err := promptForPasswordAndDeriveKey(salt)
		if err != nil {
			f.Session.Clear()
			return err
		}
		vaultKey = promptedKey
		createSession = true
	}

	// Unlock vault using the key (vault package responsibility)
	v := vault.UnlockWithKey(f.Crypto, salt, vaultKey)

	// Verify the password is correct (vault package responsibility)
	if err := vault.VerifyVaultPassword(f.System, v); err != nil {
		v.Lock()
		f.Session.Clear()
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Update factory state (command layer responsibility)
	f.Vault = v
	f.Repo.SetVault(v)
	f.Secrets = f.Repo.NewEncryptedRepository(f.Config.SecretsBucket)

	// Create new session if we prompted for password
	if createSession {
		if err := f.Session.CreateSession(vaultKey); err != nil {
			f.Logger.Error("Failed to create session: %v", err)
		}
	}

	return nil
}

// promptForPasswordAndDeriveKey prompts the user for password and derives the vault key
func promptForPasswordAndDeriveKey(salt []byte) ([]byte, error) {
	password, err := promptForPassword()
	if err != nil {
		return nil, err
	}

	key := crypto.DeriveKey(password, salt)
	return key, nil
}

// promptForPassword prompts for password with hidden input
func promptForPassword() (string, error) {
	fmt.Print("Enter master password: ")
	pwd, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	fmt.Println()
	return string(pwd), nil
}
