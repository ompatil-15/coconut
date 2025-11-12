package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/ompatil-15/coconut/internal/factory"
	"github.com/spf13/cobra"
)

func NewRootCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "coconut",
		Short: "A CLI password manager with Zero Knowledge Architecture",
		Long: `coconut is a CLI password manager with Zero Knowledge Architecture.

Store all your passwords effortlessly while having to only remember a
single master password. With the Zero Knowledge Architecture, your 
passwords are safe even after full device compromise.`,
	}

	// Vault management commands
	cmd.AddCommand(NewInitCmd(f))
	cmd.AddCommand(NewUnlockCmd(f))
	cmd.AddCommand(NewLockCmd(f))

	// Secret management commands
	cmd.AddCommand(NewAddCmd(f))
	cmd.AddCommand(NewGetCmd(f))
	cmd.AddCommand(NewListCmd(f))
	cmd.AddCommand(NewUpdateCmd(f))
	cmd.AddCommand(NewDeleteCmd(f))

	// Utility commands
	cmd.AddCommand(NewGenerateCmd(f))

	// Configuration commands
	cmd.AddCommand(NewConfigCmd(f))

	// Version command
	cmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("coconut v1.0.0")
		},
	})

	return cmd
}

func Execute() {
	cmdFactory, err := factory.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize factory: %v\n", err)
		os.Exit(1)
	}
	defer cmdFactory.Close()

	w := cmdFactory.IO.ErrOut
	logger := cmdFactory.Logger

	defer func() {
		if r := recover(); r != nil {
			if logger != nil {
				logger.Error("panic recovered: %v\n%s", r, debug.Stack())
			}
			fmt.Fprintln(w, "An unexpected error occurred. Please check the log file for details.")
			os.Exit(1)
		}
	}()

	rootCmd := NewRootCmd(cmdFactory)

	if err := rootCmd.Execute(); err != nil {
		logger.Error("Command execution failed: %v", err)
		fmt.Fprintln(w, "Error: something went wrong. Please check the log file for details.")
		os.Exit(1)
	}
}
