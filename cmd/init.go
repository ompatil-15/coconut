package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/ompatil-15/coconut/internal/config"
	"github.com/ompatil-15/coconut/internal/crypto"
	"github.com/ompatil-15/coconut/internal/db"
	"github.com/ompatil-15/coconut/internal/factory"
	"github.com/ompatil-15/coconut/internal/logger"
	"github.com/ompatil-15/coconut/internal/vault"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func NewInitCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "init",
		Aliases: []string{"initialize"},
		Short:   "Initialize a new vault (one-time setup)",
		Long: `Initialize a new vault for storing secrets. This is a one-time operation.

If you already have a vault, use 'coconut unlock' to unlock it.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return InitializeVault(f.System, f.Logger)
		},
	}

	return cmd
}

// InitializeVault creates a new vault (one-time operation)
// Returns error if vault already exists
func InitializeVault(systemRepo db.Repository, log *logger.Logger) error {
	const saltKey = "salt"

	// Check if vault already exists
	existingSalt, _ := systemRepo.Get(saltKey)
	if len(existingSalt) > 0 {
		fmt.Println("Error: Vault already exists")
		fmt.Println("")
		fmt.Println("To unlock your existing vault, use:")
		fmt.Println("  coconut unlock")
		fmt.Println("")
		fmt.Println("To lock your vault, use:")
		fmt.Println("  coconut lock")
		return fmt.Errorf("vault already initialized")
	}

	// Create new vault
	fmt.Println("Creating a new vault...")
	fmt.Println("")
	fmt.Println("Please create a strong master password:")
	fmt.Println("  • Minimum 12 characters recommended")
	fmt.Println("  • Mix of letters, numbers, and symbols")
	fmt.Println("  • Don't reuse passwords from other services")
	fmt.Println("")

	password, err := promptPasswordTwice()
	if err != nil {
		return err
	}

	// Generate salt and derive key
	salt := crypto.GenerateRandomSalt(16)
	key := crypto.DeriveKey(password, salt)

	// Create and unlock vault temporarily
	v := vault.NewVault(crypto.NewAESGCM(), salt)
	v.Unlock(key)

	// Create verification token for future password validation
	encryptedToken, err := v.CreateVerificationToken()
	if err != nil {
		return fmt.Errorf("failed to create verification token: %w", err)
	}

	// Store salt and verification token in database
	if err := systemRepo.Put(saltKey, salt); err != nil {
		return fmt.Errorf("failed to save salt: %w", err)
	}

	if err := systemRepo.Put("vault_verification", []byte(encryptedToken)); err != nil {
		return fmt.Errorf("failed to save verification token: %w", err)
	}

	if err := config.Save(systemRepo, config.Default()); err != nil {
		return fmt.Errorf("failed to save default configuration: %w", err)
	}

	log.Info("Vault initialized successfully")
	fmt.Println("")
	fmt.Println("Vault created successfully!")
	fmt.Println("")
	fmt.Println("Next steps:")
	fmt.Println("  - Add a secret:       coconut add -u username -p password")
	fmt.Println("  - List secrets:       coconut list")
	fmt.Println("  - Get a secret:       coconut get <index>")
	fmt.Println("")
	fmt.Println("Note: You'll be prompted for your master password when needed.")
	fmt.Println("")

	return nil
}

func promptPassword() (string, error) {
	pwd, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	fmt.Println()
	return string(pwd), nil
}

func promptPasswordTwice() (string, error) {
	fmt.Print("Enter password: ")
	p1, err := promptPassword()
	if err != nil {
		return "", err
	}
	fmt.Print("Confirm password: ")
	p2, err := promptPassword()
	if err != nil {
		return "", err
	}
	if p1 != p2 {
		return "", errors.New("passwords do not match")
	}
	return p1, nil
}
