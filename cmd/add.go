package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ompatil-15/coconut/internal/db/model"
	"github.com/ompatil-15/coconut/internal/factory"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func NewAddCmd(f *factory.Factory) *cobra.Command {
	var (
		username    string
		password    string
		url         string
		description string
	)

	cmd := &cobra.Command{
		Use:     "add",
		Aliases: []string{"insert"},
		Short:   "Add a new secret to the vault",
		Long:    `Adds a new secret (username, password, URL, etc.) to your encrypted vault.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := EnsureVaultUnlocked(f); err != nil {
				return err
			}

			if username == "" && password == "" && url == "" && description == "" {
				if err := readAddInteractive(f, &username, &password, &url, &description); err != nil {
					return err
				}
			}

			if username == "" {
				return fmt.Errorf("username is required")
			}
			if password == "" {
				return fmt.Errorf("password is required")
			}

			now := time.Now()
			secret := model.Secret{
				ID:          uuid.New().String(),
				Username:    username,
				Password:    password,
				URL:         url,
				Description: description,
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			if _, err := f.Secrets.Add(secret); err != nil {
				f.Logger.Error("failed to add secret: %v", err)
				return fmt.Errorf("failed to add secret: %w", err)
			}

			f.Logger.Info("Secret added successfully")
			fmt.Printf("Secret for '%s' saved successfully!\n", username)

			return nil
		},
	}

	cmd.Flags().StringVarP(&username, "username", "u", "", "Username for the secret")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password for the secret")
	cmd.Flags().StringVarP(&url, "url", "l", "", "URL for the secret")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Description for the secret")

	return cmd
}

func readAddInteractive(f *factory.Factory, username, password, url, description *string) error {
	reader := bufio.NewReader(f.IO.In)
	out := f.IO.Out

	fmt.Fprint(out, "Username: ")
	u, _ := reader.ReadString('\n')
	*username = strings.TrimSpace(u)

	fmt.Fprint(out, "Password: ")
	pwd, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(out)
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	*password = strings.TrimSpace(string(pwd))

	fmt.Fprint(out, "URL (optional): ")
	urlInput, _ := reader.ReadString('\n')
	*url = strings.TrimSpace(urlInput)

	fmt.Fprint(out, "Description (optional): ")
	desc, _ := reader.ReadString('\n')
	*description = strings.TrimSpace(desc)

	return nil
}
