package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
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

	// Setup network capture if requested
	captureNetwork := c.Bool("capture-network")
	if captureNetwork {
		fmt.Println("   ✓ Network capture enabled")
	}

	// Navigate to Ancestry.com
	if err := navigateAndSetupCapture(client, captureNetwork); err != nil {
		return err
	}

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
	fmt.Println()
	if captureNetwork {
		fmt.Println("Navigate to 'My Trees' or other pages to capture API requests...")
	}
	time.Sleep(time.Duration(waitTime) * time.Second)

	// Display and save captured network requests
	if captureNetwork {
		if err := displayCapturedRequests(client, c); err != nil {
			fmt.Printf("Warning: %v\n", err)
		}
	}

	fmt.Println()
	fmt.Println("Test completed! Closing browser...")

	return nil
}

// navigateAndSetupCapture navigates to Ancestry and sets up network capture
func navigateAndSetupCapture(client *ancestry.Client, captureEnabled bool) error {
	fmt.Println("2. Navigating to Ancestry.com...")
	if err := client.NavigateToAncestry(); err != nil {
		return fmt.Errorf("failed to navigate: %w", err)
	}
	fmt.Println("   ✓ Navigation successful")

	// Enable network capture after navigation
	if captureEnabled {
		if err := client.EnableNetworkCapture(); err != nil {
			fmt.Printf("   ✗ Warning: Failed to enable network capture: %v\n", err)
		}
	}

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

// displayCapturedRequests displays and optionally saves captured network requests
func displayCapturedRequests(client *ancestry.Client, c *cli.Context) error {
	requests := client.GetCapturedRequests()

	fmt.Println("\n=== CAPTURED NETWORK REQUESTS ===")
	fmt.Printf("Total requests captured: %d\n\n", len(requests))

	if len(requests) == 0 {
		fmt.Println("No API requests were captured.")
		fmt.Println("Try navigating to 'My Trees' or other pages to see API calls.")
		return nil
	}

	// Display each request
	for i, req := range requests {
		fmt.Printf("[%d] %s %s\n", i+1, req.Method, req.URL)
		fmt.Printf("    Status: %d\n", req.StatusCode)
		fmt.Printf("    Content-Type: %s\n", req.ContentType)
		fmt.Printf("    Timestamp: %s\n", req.Timestamp.Format("15:04:05"))

		// Show request body if present (truncated)
		if req.RequestBody != "" {
			truncated := truncateString(req.RequestBody, 100)
			fmt.Printf("    Request: %s\n", truncated)
		}

		// Show response body preview (truncated)
		if req.ResponseBody != "" {
			// Try to pretty-print JSON
			var jsonData interface{}
			if err := json.Unmarshal([]byte(req.ResponseBody), &jsonData); err == nil {
				prettyJSON, _ := json.MarshalIndent(jsonData, "    ", "  ")
				truncated := truncateString(string(prettyJSON), 200)
				fmt.Printf("    Response: %s\n", truncated)
			} else {
				truncated := truncateString(req.ResponseBody, 200)
				fmt.Printf("    Response: %s\n", truncated)
			}
		}
		fmt.Println()
	}

	// Save to file if output flag is set
	outputFile := c.String("output")
	if outputFile != "" {
		if err := saveCapturedRequests(requests, outputFile); err != nil {
			return fmt.Errorf("failed to save captured requests: %w", err)
		}
		fmt.Printf("✓ Saved %d requests to %s\n", len(requests), outputFile)
	}

	return nil
}

// saveCapturedRequests saves captured requests to a JSON file
func saveCapturedRequests(requests []*ancestry.CapturedRequest, filename string) error {
	data, err := json.MarshalIndent(requests, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal requests: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// truncateString truncates a string to maxLen and adds "..." if truncated
func truncateString(s string, maxLen int) string {
	// Remove newlines and extra spaces
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	s = strings.TrimSpace(s)

	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
