package cmd

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/atotto/clipboard"
	"github.com/ompatil-15/coconut/internal/factory"
	"github.com/spf13/cobra"
)

const (
	lowercase = "abcdefghijklmnopqrstuvwxyz"
	uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits    = "0123456789"
	special   = "!@#$%^&*()_+-=[]{}|;:,.<>?"
)

func NewGenerateCmd(f *factory.Factory) *cobra.Command {
	var (
		length int
		copy   bool
	)

	cmd := &cobra.Command{
		Use:     "generate",
		Aliases: []string{"gen"},
		Short:   "Generate a random strong password",
		Long:    `Generate a random strong password with letters, numbers, and special characters.`,
		Example: `  coconut generate
  coconut generate --length 16
  coconut generate -l 20 --copy`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if length < 4 {
				return fmt.Errorf("password length must be at least 4")
			}

			password, err := generatePassword(length)
			if err != nil {
				return fmt.Errorf("failed to generate password: %w", err)
			}

			fmt.Printf("Generated password: %s\n", password)

			if copy {
				if err := clipboard.WriteAll(password); err != nil {
					fmt.Println("Warning: Failed to copy to clipboard")
				} else {
					fmt.Println("Password copied to clipboard!")
				}
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&length, "length", "l", 16, "Password length")
	cmd.Flags().BoolVarP(&copy, "copy", "c", false, "Copy password to clipboard")

	return cmd
}

func generatePassword(length int) (string, error) {
	charset := lowercase + uppercase + digits + special
	password := make([]byte, length)

	// Ensure at least one character from each category
	password[0] = lowercase[mustRandomInt(len(lowercase))]
	password[1] = uppercase[mustRandomInt(len(uppercase))]
	password[2] = digits[mustRandomInt(len(digits))]
	password[3] = special[mustRandomInt(len(special))]

	for i := 4; i < length; i++ {
		password[i] = charset[mustRandomInt(len(charset))]
	}

	for i := length - 1; i > 0; i-- {
		j := mustRandomInt(i + 1)
		password[i], password[j] = password[j], password[i]
	}

	return string(password), nil
}

func mustRandomInt(max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		panic(err)
	}
	return int(n.Int64())
}
