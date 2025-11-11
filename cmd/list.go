package cmd

import (
	"fmt"
	"strings"

	"github.com/ompatil-15/coconut/internal/factory"
	"github.com/spf13/cobra"
)

var verbose bool

func NewListCmd(f *factory.Factory) *cobra.Command {
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "View all your saved secrets securely",
		Long: `Retrieves and displays all secret entries from the encrypted vault. 
By default, only essential metadata is shown. Use --verbose for detailed view.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := EnsureVaultUnlocked(f); err != nil {
				return err
			}

			out := f.IO.Out
			errOut := f.IO.ErrOut
			logger := f.Logger

			logger.Info("Executing 'list' command (verbose=%v)", verbose)

			secrets, err := f.Secrets.List()
			if err != nil {
				logger.Error("Failed to fetch secrets: %v", err)
				fmt.Fprintf(errOut, "Error: failed to fetch secrets: %v\n", err)
				return fmt.Errorf("failed to fetch secrets: %w", err)
			}

			if len(secrets) == 0 {
				fmt.Fprintln(out, "No secrets found in the vault.")
				logger.Info("No secrets found in vault")
				return nil
			}

			logger.Info("Fetched %d secrets from vault", len(secrets))

			var headerFmt, rowFmt, divider string
			if verbose {
				headerFmt = "%-10s %-30s %-30s %-15s %s\n"
				rowFmt = "%-10d %-30s %-30s %-15s %s\n"
				divider = strings.Repeat("-", 120)
			} else {
				headerFmt = "%-10s %-30s %-30s %s\n"
				rowFmt = "%-10d %-30s %-30s %s\n"
				divider = strings.Repeat("-", 100)
			}

			if verbose {
				fmt.Fprintf(out, headerFmt, "ID", "USERNAME", "URL", "CREATED", "DESCRIPTION")
			} else {
				fmt.Fprintf(out, headerFmt, "ID", "USERNAME", "URL", "DESCRIPTION")
			}
			fmt.Fprintln(out, divider)

			for i, secret := range secrets {
				if verbose {
					fmt.Fprintf(out, rowFmt,
						i+1,
						truncate(secret.Username, 20),
						truncate(secret.URL, 40),
						secret.CreatedAt.Format("2006-01-02"),
						truncate(secret.Description, 50),
					)
				} else {
					fmt.Fprintf(out, rowFmt,
						i+1,
						truncate(secret.Username, 20),
						truncate(secret.URL, 40),
						truncate(secret.Description, 50),
					)
				}
			}

			logger.Info("Successfully listed all secrets")
			return nil
		},
	}

	listCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")
	return listCmd
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
