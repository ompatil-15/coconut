package cmd

import (
	"fmt"

	"github.com/ompatil-15/coconut/internal/factory"
	"github.com/spf13/cobra"
)

func NewLockCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock",
		Short: "Lock the vault",
		Long: `Lock your vault to secure your secrets.

After locking, you'll need to run 'coconut unlock' and enter your 
master password again to access your secrets.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Clear the session (removes cached key)
			if err := f.Session.Clear(); err != nil {
				f.Logger.Error("Failed to clear session: %v", err)
			}

			// Lock the vault in memory if it's unlocked
			if f.Vault != nil && f.Vault.IsUnlocked() {
				f.Vault.Lock()
			}

			f.Logger.Info("Vault locked and session cleared")

			fmt.Println("Vault locked successfully!")
			fmt.Println("")
			fmt.Println("Your session has been cleared.")
			fmt.Println("You'll need to enter your master password again for the next operation.")
			fmt.Println("")

			return nil
		},
	}

	return cmd
}
