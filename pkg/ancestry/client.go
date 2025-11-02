package ancestry

import (
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

const (
	// AncestryBaseURL is the main Ancestry.com URL
	AncestryBaseURL = "https://www.ancestry.com"
)

// Client represents a browser automation client for Ancestry.com
type Client struct {
	browser *rod.Browser
	page    *rod.Page
}

// NewClient creates a new Ancestry client with browser automation
func NewClient() (*Client, error) {
	// Launch browser with visible window (not headless)
	// Use launcher to handle browser binary detection and launching
	path, found := launcher.LookPath()
	if !found {
		return nil, fmt.Errorf("browser not found: please install Chrome or Chromium")
	}

	u := launcher.New().Bin(path).Headless(false).MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect()

	return &Client{
		browser: browser,
	}, nil
}

// NavigateToAncestry opens Ancestry.com in the browser
func (c *Client) NavigateToAncestry() error {
	// Create a new page
	page := c.browser.MustPage()
	c.page = page

	// Navigate to Ancestry.com
	if err := c.page.Navigate(AncestryBaseURL); err != nil {
		return fmt.Errorf("failed to navigate to Ancestry.com: %w", err)
	}

	// Wait for the page to load
	if err := c.page.WaitLoad(); err != nil {
		return fmt.Errorf("failed to wait for page load: %w", err)
	}

	// Add a small delay to ensure page is fully rendered
	time.Sleep(2 * time.Second)

	return nil
}

// Close closes the browser and cleans up resources
func (c *Client) Close() error {
	if c.page != nil {
		if err := c.page.Close(); err != nil {
			return fmt.Errorf("failed to close page: %w", err)
		}
	}

	if c.browser != nil {
		if err := c.browser.Close(); err != nil {
			return fmt.Errorf("failed to close browser: %w", err)
		}
	}

	return nil
}

// GetPage returns the current page instance
func (c *Client) GetPage() *rod.Page {
	return c.page
}
