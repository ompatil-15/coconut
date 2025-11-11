package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ompatil-15/coconut/internal/db/model"
	"github.com/ompatil-15/coconut/internal/factory"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func NewUpdateCmd(f *factory.Factory) *cobra.Command {
	var (
		username    string
		url         string
		description string
	)

	cmd := &cobra.Command{
		Use:     "update <index> [--username USERNAME] [--url URL] [--description DESCRIPTION]",
		Aliases: []string{"edit"},
		Short:   "Update one or more fields of a secret",
		Long: `Update stored secrets securely. 
Only provided fields are changed; others remain unchanged.
If no flags are given, the command will prompt interactively.`,

		Example: `
  coconut update 3
  coconut update 2 --username "new_user" --url "https://coconut.pm"
  coconut update 1 --username "admin"`,

		Args: cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			if err := EnsureVaultUnlocked(f); err != nil {
				return err
			}

			index, err := strconv.Atoi(args[0])
			if err != nil {
				return errors.New("please provide a valid index number (e.g. 1, 2, 3)")
			}

			secrets, err := f.Secrets.List()
			if err != nil {
				return fmt.Errorf("failed to fetch secrets: %w", err)
			}
			if index < 1 || index > len(secrets) {
				return fmt.Errorf("invalid index: %d (valid range: 1–%d)", index, len(secrets))
			}

			secret := secrets[index-1]

			if username == "" && url == "" && description == "" {
				if err := readInteractive(f, &secret); err != nil {
					return err
				}
			} else {
				if username != "" {
					secret.Username = username
				}
				if url != "" {
					secret.URL = url
				}
				if description != "" {
					secret.Description = description
				}
			}

			if err := f.Secrets.Update(secret); err != nil {
				return fmt.Errorf("failed to update secret: %w", err)
			}

			fmt.Printf("Secret with id %d updated successfully.\n", index)
			return nil
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "New username")
	cmd.Flags().StringVar(&url, "url", "", "New URL")
	cmd.Flags().StringVar(&description, "description", "", "New description")

	return cmd
}

func readInteractive(f *factory.Factory, secret *model.Secret) error {
	reader := bufio.NewReader(f.IO.In)
	out := f.IO.Out

	fmt.Fprintf(out, "Username (leave blank to keep '%s'): ", secret.Username)
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username != "" {
		secret.Username = username
	}

	fmt.Fprint(out, "Password (leave blank to keep current): ")
	pwd, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(out)
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	password := strings.TrimSpace(string(pwd))
	if password != "" {
		secret.Password = password
	}

	fmt.Fprintf(out, "URL (leave blank to keep '%s'): ", secret.URL)
	url, _ := reader.ReadString('\n')
	url = strings.TrimSpace(url)
	if url != "" {
		secret.URL = url
	}

	fmt.Fprintf(out, "Description (leave blank to keep '%s'): ", secret.Description)
	desc, _ := reader.ReadString('\n')
	desc = strings.TrimSpace(desc)
	if desc != "" {
		secret.Description = desc
	}

	return nil
}
