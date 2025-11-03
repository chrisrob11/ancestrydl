package commands

import (
	"fmt"
	"time"

	"github.com/chrisrob11/ancestrydl/pkg/ancestry"
	"github.com/urfave/cli/v2"
)

// TestBrowser tests the browser automation by opening Ancestry.com
func TestBrowser(c *cli.Context) error {
	username := c.String("username")
	password := c.String("password")
	testLogin := username != "" && password != ""

	fmt.Println("Testing browser automation...")
	if testLogin {
		fmt.Println("This will open a browser, navigate to Ancestry.com, and test login")
	} else {
		fmt.Println("This will open a browser window and navigate to Ancestry.com")
	}
	fmt.Println()

	// Create a new client
	fmt.Println("1. Creating browser client...")
	client, err := ancestry.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			fmt.Printf("Warning: failed to close browser: %v\n", err)
		}
	}()
	fmt.Println("   ✓ Browser launched successfully")

	// Navigate to Ancestry.com
	fmt.Println("2. Navigating to Ancestry.com...")
	if err := client.NavigateToAncestry(); err != nil {
		return fmt.Errorf("failed to navigate: %w", err)
	}
	fmt.Println("   ✓ Navigation successful")

	// Get page info
	page := client.GetPage()
	if page != nil {
		info := page.MustInfo()
		fmt.Printf("   ✓ Current URL: %s\n", info.URL)
		fmt.Printf("   ✓ Page title: %s\n", info.Title)
	}

	// Test login if credentials provided
	if testLogin {
		if err := testLoginFlow(client, c); err != nil {
			return err
		}
	}

	// Keep browser open so you can see it
	waitTime := 10
	if testLogin {
		waitTime = 15
	}
	fmt.Println()
	fmt.Printf("Browser will stay open for %d seconds so you can see it...\n", waitTime)
	fmt.Println("Press Ctrl+C to close early")
	time.Sleep(time.Duration(waitTime) * time.Second)

	fmt.Println()
	fmt.Println("Test completed! Closing browser...")

	return nil
}

// testLoginFlow handles the login testing logic
func testLoginFlow(client *ancestry.Client, c *cli.Context) error {
	username := c.String("username")
	password := c.String("password")
	noSubmit := c.Bool("no-submit")
	twoFactorMethod := c.String("2fa")

	if noSubmit {
		fmt.Println("3. Testing login (filling fields only, NOT submitting)...")
	} else {
		fmt.Println("3. Testing login...")
		if twoFactorMethod != "" {
			fmt.Printf("   2FA method: %s\n", twoFactorMethod)
		}
	}

	if err := client.LoginWithOptions(username, password, ancestry.LoginOptions{
		SkipSubmit:      noSubmit,
		TwoFactorMethod: twoFactorMethod,
	}); err != nil {
		fmt.Println("   ✗ Login failed!")
		return fmt.Errorf("authentication failed: %w", err)
	}

	if noSubmit {
		fmt.Println("   ✓ Form filled successfully (NOT submitted)")
		fmt.Println("   → Check the browser to verify username and password are correct")
		return nil
	}

	return handleSuccessfulLogin(client)
}

// handleSuccessfulLogin handles post-login tasks including cookie extraction
func handleSuccessfulLogin(client *ancestry.Client) error {
	fmt.Println("   ✓ Login successful!")

	// Verify logged in state
	if client.IsLoggedIn() {
		fmt.Println("   ✓ User is authenticated")
	}

	// Show current page after login
	page := client.GetPage()
	if page != nil {
		info := page.MustInfo()
		fmt.Printf("   ✓ Current URL: %s\n", info.URL)
	}

	// Extract and display cookies
	return extractAndDisplayCookies(client)
}

// extractAndDisplayCookies extracts session cookies and displays information
func extractAndDisplayCookies(client *ancestry.Client) error {
	fmt.Println("\n4. Extracting session cookies...")
	cookies, err := client.GetAncestrySessionCookies()
	if err != nil {
		fmt.Printf("   ✗ Failed to extract cookies: %v\n", err)
		return nil // Non-fatal error
	}

	fmt.Printf("   ✓ Extracted %d cookie(s) from Ancestry.com\n", len(cookies))

	// Display cookie names (not values for security)
	for i, cookie := range cookies {
		fmt.Printf("      [%d] %s (domain: %s, secure: %t, httpOnly: %t)\n",
			i+1, cookie.Name, cookie.Domain, cookie.Secure, cookie.HTTPOnly)
	}

	// Serialize cookies for storage
	cookiesJSON, err := ancestry.SerializeCookies(cookies)
	if err != nil {
		fmt.Printf("   ✗ Failed to serialize cookies: %v\n", err)
		return nil // Non-fatal error
	}

	fmt.Printf("   ✓ Cookies serialized (%d bytes)\n", len(cookiesJSON))
	fmt.Println("   → These cookies can be used for HTTP API requests")

	return nil
}
