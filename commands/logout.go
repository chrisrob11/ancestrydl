package commands

import (
	"fmt"

	"github.com/chrisrob11/ancestrydl/pkg/config"
	"github.com/urfave/cli/v2"
)

// Logout handles the logout command, removing stored credentials from the system keyring
func Logout(c *cli.Context) error {
	// Check if credentials exist before attempting to delete
	creds, err := config.GetCredentials()
	if err != nil {
		if err == config.ErrCredentialsNotFound {
			fmt.Println("No credentials found. You are not logged in.")
			return nil
		}
		return fmt.Errorf("failed to check credentials: %w", err)
	}

	// Delete credentials from keyring
	if err := config.DeleteCredentials(); err != nil {
		return fmt.Errorf("failed to remove credentials: %w", err)
	}

	fmt.Printf("Successfully logged out user: %s\n", creds.Username)
	fmt.Println("Your credentials have been removed from the system keyring.")

	return nil
}
