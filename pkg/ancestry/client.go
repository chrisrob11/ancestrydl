package ancestry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
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

// GetCookies retrieves all cookies from the current browser session
func (c *Client) GetCookies() ([]*proto.NetworkCookie, error) {
	if c.page == nil {
		return nil, fmt.Errorf("no page available")
	}

	cookies, err := c.page.Cookies([]string{})
	if err != nil {
		return nil, fmt.Errorf("failed to get cookies: %w", err)
	}

	return cookies, nil
}

// GetAncestrySessionCookies retrieves only cookies relevant to Ancestry.com authentication
func (c *Client) GetAncestrySessionCookies() ([]*proto.NetworkCookie, error) {
	allCookies, err := c.GetCookies()
	if err != nil {
		return nil, err
	}

	var ancestryCookies []*proto.NetworkCookie
	for _, cookie := range allCookies {
		// Filter for ancestry.com domain cookies
		if cookie.Domain == ".ancestry.com" || cookie.Domain == "www.ancestry.com" || cookie.Domain == "ancestry.com" {
			ancestryCookies = append(ancestryCookies, cookie)
		}
	}

	return ancestryCookies, nil
}

// CookiesToHTTPCookies converts rod cookies to standard http.Cookie format
func CookiesToHTTPCookies(rodCookies []*proto.NetworkCookie) []*http.Cookie {
	httpCookies := make([]*http.Cookie, len(rodCookies))
	for i, rc := range rodCookies {
		httpCookies[i] = &http.Cookie{
			Name:     rc.Name,
			Value:    rc.Value,
			Path:     rc.Path,
			Domain:   rc.Domain,
			Expires:  time.Unix(int64(rc.Expires), 0),
			Secure:   rc.Secure,
			HttpOnly: rc.HTTPOnly,
			SameSite: http.SameSiteDefaultMode,
		}
	}
	return httpCookies
}

// SerializeCookies converts cookies to JSON for storage
func SerializeCookies(cookies []*proto.NetworkCookie) (string, error) {
	data, err := json.Marshal(cookies)
	if err != nil {
		return "", fmt.Errorf("failed to serialize cookies: %w", err)
	}
	return string(data), nil
}

// DeserializeCookies converts JSON back to cookies
func DeserializeCookies(data string) ([]*proto.NetworkCookie, error) {
	var cookies []*proto.NetworkCookie
	if err := json.Unmarshal([]byte(data), &cookies); err != nil {
		return nil, fmt.Errorf("failed to deserialize cookies: %w", err)
	}
	return cookies, nil
}
