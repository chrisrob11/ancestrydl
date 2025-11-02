package commands

import (
	"fmt"
	"strings"

	"github.com/chrisrob11/ancestrydl/pkg/config"
	"github.com/urfave/cli/v2"
)

// Login handles the login command, storing user credentials in the system keyring
func Login(c *cli.Context) error {
	username := strings.TrimSpace(c.String("username"))
	password := c.String("password")

	// Validate inputs
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	// Save credentials to keyring
	if err := config.SaveCredentials(username, password); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	fmt.Printf("Successfully saved credentials for user: %s\n", username)
	fmt.Println("You can now use 'ancestrydl list-trees' and 'ancestrydl export' commands.")

	return nil
}
