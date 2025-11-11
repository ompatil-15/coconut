package cmd

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/ompatil-15/coconut/internal/factory"
	"github.com/spf13/cobra"
)

func NewDeleteCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <index>",
		Aliases: []string{"del", "rm"},
		Short:   "Delete a saved secret from the vault",
		Long:    `Safely deletes a specific secret from your encrypted vault using its index.`,
		Args:    cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			if err := EnsureVaultUnlocked(f); err != nil {
				return err
			}

			out := f.IO.Out
			errOut := f.IO.ErrOut
			logger := f.Logger

			index, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Fprintln(errOut, "Error: please provide a valid index number (e.g. 1, 2, 3)")
				return nil
			}

			secrets, err := f.Secrets.List()
			if err != nil {
				logger.Error("Failed to list secrets: %v", err)
				fmt.Fprintln(errOut, "Error: failed to fetch secrets. Check log for details.")
				return err
			}

			if index < 1 || index > len(secrets) {
				fmt.Fprintf(errOut, "Error: invalid index %d (valid range: 1–%d)\n", index, len(secrets))
				return nil
			}

			secret := secrets[index-1]
			reader := bufio.NewReader(f.IO.In)
			fmt.Fprintf(out, "Are you sure you want to delete secret %d (%s)? (y/N): ", index, secret.Username)

			confirm, _ := reader.ReadString('\n')
			confirm = strings.TrimSpace(confirm)

			if strings.ToLower(confirm) != "y" {
				fmt.Fprintln(out, "Delete cancelled.")
				logger.Info("Delete cancelled for secret %d", index)
				return nil
			}

			if err := f.Secrets.Delete(secret.ID); err != nil {
				logger.Error("Failed to delete secret %d: %v", index, err)
				fmt.Fprintln(errOut, "Error: failed to delete secret. Check log for details.")
				return err
			}

			fmt.Fprintf(out, "Secret %d deleted successfully.\n", index)
			logger.Info("Secret %d deleted successfully", index)
			return nil
		},
	}

	return cmd
}
