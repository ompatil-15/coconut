package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/ompatil-15/coconut/internal/db/model"
	"github.com/ompatil-15/coconut/internal/factory"
	"github.com/spf13/cobra"
)

func NewGetCmd(f *factory.Factory) *cobra.Command {
	var (
		showPassword bool
		copyToClip   bool
	)

	cmd := &cobra.Command{
		Use:   "get <index>",
		Short: "Retrieve a specific secret from the vault",
		Long: `Fetch details of a single secret from the encrypted vault using its index 
(as shown in the list command). By default, the password is hidden. 

Use:
  - '--show-password' or '-s' to reveal the password in terminal
  - '--copy' or '-c' to copy the password to clipboard silently.`,
		Example: `coconut get <index>
coconut get <index> -c
coconut get <index> -s`,
		Args: cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			// Ensure vault is unlocked
			if err := EnsureVaultUnlocked(f); err != nil {
				return err
			}

			index, err := strconv.Atoi(args[0])
			if err != nil {
				return errors.New("please provide a valid index number (e.g. 1, 2, 3)")
			}

			secrets, err := f.Secrets.List()
			if err != nil {
				f.Logger.Error("failed to fetch secrets: %v", err)
				return fmt.Errorf("failed to fetch secrets: %w", err)
			}

			if index < 1 || index > len(secrets) {
				return fmt.Errorf("invalid index: %d (valid range: 1–%d)", index, len(secrets))
			}

			secret := secrets[index-1]

			if copyToClip {
				if err := clipboard.WriteAll(secret.Password); err != nil {
					f.Logger.Error("failed to copy password: %v", err)
					return fmt.Errorf("failed to copy password to clipboard: %w", err)
				}
				fmt.Println("Password copied to clipboard securely.")
				return nil
			}

			displaySecret(&secret, showPassword)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&showPassword, "show-password", "s", false, "Show the password value explicitly")
	cmd.Flags().BoolVarP(&copyToClip, "copy", "c", false, "Copy the password to clipboard without showing it")

	return cmd
}

func displaySecret(secret *model.Secret, reveal bool) {
	// fmt.Printf("%-15s: %s\n", "ID", secret.ID)
	fmt.Printf("%-15s: %s\n", "Username", secret.Username)

	if reveal {
		fmt.Printf("%-15s: %s\n", "Password", secret.Password)
	} else {
		fmt.Printf("%-15s: %s\n", "Password", maskPassword(secret.Password))
	}

	fmt.Printf("%-15s: %s\n", "URL", secret.URL)
	fmt.Printf("%-15s: %s\n", "Description", secret.Description)
	fmt.Printf("%-15s: %s\n", "Created At", secret.CreatedAt.Format("2006-01-02 15:04"))
	fmt.Printf("%-15s: %s\n", "Updated At", secret.UpdatedAt.Format("2006-01-02 15:04"))
}

func maskPassword(pw string) string {
	if len(pw) == 0 {
		return "-"
	}
	return strings.Repeat("*", 8)
}
