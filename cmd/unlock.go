package cmd

import (
	"fmt"

	"github.com/ompatil-15/coconut/internal/factory"
	"github.com/spf13/cobra"
)

func NewUnlockCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unlock",
		Short: "Pre-unlock the vault with your master password (optional)",
		Long: `Pre-unlock your vault with your master password.

Note: This command is optional. All secret management commands will 
automatically prompt for your master password if the vault is locked.

Use 'coconut lock' to lock the vault when done.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if already unlocked
			if f.Vault != nil && f.Vault.IsUnlocked() && f.Session.IsValid() {
				fmt.Println("Vault is already unlocked")
				return nil
			}

			// Use centralized unlock logic
			if err := EnsureVaultUnlocked(f); err != nil {
				return err
			}

			remaining := f.Session.GetRemainingTime()
			minutes := int(remaining.Minutes())

			f.Logger.Info("Vault unlocked successfully with session")
			fmt.Println("Vault unlocked successfully!")
			fmt.Println("")
			fmt.Printf("Session created: You won't need to re-enter your password for %d minutes.\n", minutes)
			fmt.Println("")
			fmt.Println("You can now:")
			fmt.Println("  - Add secrets:    coconut add -u username -p password")
			fmt.Println("  - List secrets:   coconut list")
			fmt.Println("  - Get secrets:    coconut get <index>")
			fmt.Println("  - Lock vault:     coconut lock")
			fmt.Println("")

			return nil
		},
	}

	return cmd
}


