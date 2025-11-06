package commands

import (
	"fmt"
	"strings"

	"github.com/chrisrob11/ancestrydl/pkg/ancestry"
	"github.com/chrisrob11/ancestrydl/pkg/config"
	"github.com/urfave/cli/v2"
)

// Login handles the login command using browser automation to authenticate and extract cookies
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

	fmt.Println("Starting authentication process...")
	fmt.Println("This will open a browser window to log you in securely.")
	fmt.Println()

	// Create browser client
	fmt.Println("1. Launching browser...")
	client, err := ancestry.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create browser client: %w", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			fmt.Printf("Warning: failed to close browser: %v\n", err)
		}
	}()
	fmt.Println("   ✓ Browser launched")

	// Navigate to Ancestry
	fmt.Println("2. Navigating to Ancestry.com...")
	if err := client.NavigateToAncestry(); err != nil {
		return fmt.Errorf("failed to navigate to Ancestry.com: %w", err)
	}
	fmt.Println("   ✓ Navigation complete")

	// Perform login
	fmt.Println("3. Logging in...")
	loginOpts := ancestry.LoginOptions{}

	// Check if 2FA method was specified
	twoFactorMethod := c.String("2fa")
	if twoFactorMethod != "" {
		loginOpts.TwoFactorMethod = twoFactorMethod
		fmt.Printf("   Using 2FA method: %s\n", twoFactorMethod)
	}

	if err := client.LoginWithOptions(username, password, loginOpts); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	fmt.Println("   ✓ Login successful")

	// Extract cookies
	fmt.Println("4. Extracting session cookies...")
	cookies, err := client.GetAncestrySessionCookies()
	if err != nil {
		return fmt.Errorf("failed to extract cookies: %w", err)
	}
	fmt.Printf("   ✓ Extracted %d cookies\n", len(cookies))

	// Serialize cookies to JSON
	cookiesJSON, err := ancestry.SerializeCookies(cookies)
	if err != nil {
		return fmt.Errorf("failed to serialize cookies: %w", err)
	}

	// Save cookies to keyring
	fmt.Println("5. Saving session to keyring...")
	if err := config.SaveCookies(cookiesJSON); err != nil {
		return fmt.Errorf("failed to save cookies: %w", err)
	}
	fmt.Println("   ✓ Session saved")

	// Also save credentials for reference
	if err := config.SaveCredentials(username, password); err != nil {
		// Don't fail if we can't save credentials, cookies are more important
		fmt.Printf("   Warning: failed to save credentials: %v\n", err)
	}

	fmt.Println()
	fmt.Println("✅ Authentication completed successfully!")
	fmt.Printf("   Logged in as: %s\n", username)
	fmt.Println()
	fmt.Println("You can now use commands like:")
	fmt.Println("  • ancestrydl list-trees")
	fmt.Println("  • ancestrydl list-people <tree-id>")
	fmt.Println()
	fmt.Println("Browser will close in 3 seconds...")

	// Keep browser open briefly so user can see success
	return nil
}
