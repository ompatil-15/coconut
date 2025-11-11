package cmd

import (
	"fmt"
	"strconv"

	"github.com/ompatil-15/coconut/internal/config"
	"github.com/ompatil-15/coconut/internal/factory"
	"github.com/spf13/cobra"
)

func NewConfigCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configure coconut settings",
		Long:  `View and modify coconut configuration settings.`,
	}

	cmd.AddCommand(newConfigGetCmd(f))
	cmd.AddCommand(newConfigSetCmd(f))

	return cmd
}

func newConfigGetCmd(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <setting>",
		Short: "Get a configuration value",
		Long: `Get the current value of a configuration setting.

Available settings:
  autolock    Inactivity timeout in seconds before autolocking (default: 300)`,
		Example: `coconut config get autolock`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			setting := args[0]

			switch setting {
			case "autolock":
				timeout := getAutoLockTimeout(f)
				minutes := float64(timeout) / 60.0
				fmt.Printf("Autolock timeout: %d seconds (%.2f minutes)\n", timeout, minutes)
				return nil
			default:
				return fmt.Errorf("unknown setting: %s\nAvailable settings: autolock", setting)
			}
		},
	}
}

func newConfigSetCmd(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "set <setting> <value>",
		Short: "Set a configuration value",
		Long: `Set the value of a configuration setting.

Available settings:
  autolock    Inactivity timeout in seconds before autolocking
              The vault locks after this many seconds of no command activity.
              Each command execution resets the inactivity timer.

Examples:
  0    = autolock disabled
  300  = 5 minutes of inactivity (default)
  600  = 10 minutes of inactivity
  900  = 15 minutes of inactivity
  1800 = 30 minutes of inactivity
  3600 = 1 hour of inactivity`,
		Example: `coconut config set autolock 600`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			setting := args[0]
			value := args[1]

			switch setting {
			case "autolock":
				seconds, err := strconv.Atoi(value)
				if err != nil {
					return fmt.Errorf("invalid value: must be a number (seconds)")
				}

				if seconds > 86400 {
					return fmt.Errorf("autolock timeout must be at most 86400 seconds (24 hours)")
				}

				if err := setAutoLockTimeout(f, seconds); err != nil {
					return fmt.Errorf("failed to set autolock timeout: %w", err)
				}

				minutes := float64(seconds) / 60.0
				if seconds == 0 {
					fmt.Println("Autolock disabled: Vault will remain unlocked until manually locked.")
				} else {
					fmt.Printf("Autolock set to %d seconds (%.2f minutes) of inactivity\n", seconds, minutes)
				}
				fmt.Println("")
				fmt.Println("Note: This will take effect on your next unlock.")
				fmt.Println("Current session will continue with the previous timeout.")

				f.Logger.Info("Autolock timeout changed to %d seconds", seconds)
				return nil

			default:
				return fmt.Errorf("unknown setting: %s\nAvailable settings: autolock", setting)
			}
		},
	}
}

func getAutoLockTimeout(f *factory.Factory) int {
	return f.Config.AutoLockSecs
}

func setAutoLockTimeout(f *factory.Factory, seconds int) error {
	f.Config.AutoLockSecs = seconds
	return config.Save(f.System, f.Config)
}
