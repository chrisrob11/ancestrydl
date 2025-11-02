package commands

import (
	"fmt"
	"time"

	"github.com/chrisrob11/ancestrydl/pkg/ancestry"
	"github.com/urfave/cli/v2"
)

// TestBrowser tests the browser automation by opening Ancestry.com
func TestBrowser(c *cli.Context) error {
	fmt.Println("Testing browser automation...")
	fmt.Println("This will open a browser window and navigate to Ancestry.com")
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

	// Keep browser open so you can see it
	fmt.Println()
	fmt.Println("Browser will stay open for 10 seconds so you can see it...")
	fmt.Println("Press Ctrl+C to close early")
	time.Sleep(10 * time.Second)

	fmt.Println()
	fmt.Println("Test completed! Closing browser...")

	return nil
}
