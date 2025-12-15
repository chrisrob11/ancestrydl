package ancestry

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/go-rod/rod/lib/proto"
	"golang.org/x/net/publicsuffix"
)

// APIClient handles HTTP requests to Ancestry.com APIs
type APIClient struct {
	httpClient       *http.Client
	baseURL          string
	loggingTransport *loggingTransport // For verbose mode
	userID           string            // Added: Stores the authenticated user's ID
	log              *log.Logger       // Added: Logger for client-specific messages
}

// NewAPIClient creates a new API client with the given cookies
func NewAPIClient(cookies []*proto.NetworkCookie, verbose bool) (*APIClient, error) {
	// Initialize logger
	clientLogger := log.New(os.Stderr, "[APIClient] ", log.LstdFlags)
	if !verbose {
		clientLogger.SetOutput(io.Discard) // Suppress output if not verbose
	}

	// Create cookie jar
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	// Convert rod cookies to http.Cookie and add to jar
	ancestryURL, _ := url.Parse("https://www.ancestry.com")
	httpCookies := CookiesToHTTPCookies(cookies)
	jar.SetCookies(ancestryURL, httpCookies)

	// Attempt to extract userID from cookies, e.g., from s_vi cookie
	extractedUserID := ""
	for _, cookie := range httpCookies {
		if cookie.Name == "s_vi" {
			// s_vi cookie often contains the user ID like "s_vi=[CS]v1|2B452D64051D56F5-60000108A000676C[CE]"
			// We need to parse this out carefully. This is a heuristic.
			if matches := regexp.MustCompile(`v1\|(.*?)\|`).FindStringSubmatch(cookie.Value); len(matches) > 1 {
				extractedUserID = matches[1]
				break
			}
		}
		// Another common place is a "AMCV_###@AdobeOrg" cookie, but s_vi is often more reliable
	}

	// Create base HTTP transport
	baseTransport := http.DefaultTransport

	// Wrap with logging transport if verbose mode is enabled
	var logTransport *loggingTransport
	var finalTransport http.RoundTripper = baseTransport
	if verbose {
		logTransport, err = newLoggingTransport(baseTransport)
		if err != nil {
			return nil, fmt.Errorf("failed to create logging transport: %w", err)
		}
		finalTransport = logTransport
	}

	client := &http.Client{
		Jar:       jar,
		Timeout:   30 * time.Second,
		Transport: finalTransport,
	}

	return &APIClient{
		httpClient:       client,
		baseURL:          "https://www.ancestry.com",
		loggingTransport: logTransport,
		userID:           extractedUserID, // Initialized userID
		log:              clientLogger,    // Initialized logger
	}, nil
}

// NewAPIClientFromJSON creates an API client from serialized JSON cookies
func NewAPIClientFromJSON(cookiesJSON string, verbose bool) (*APIClient, error) {
	cookies, err := DeserializeCookies(cookiesJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize cookies: %w", err)
	}
	return NewAPIClient(cookies, verbose)
}

// GetUserID retrieves the authenticated user's ID, fetching it if not already known.
func (c *APIClient) GetUserID() (string, error) {
	if c.userID != "" {
		return c.userID, nil
	}

	c.log.Println("Attempting to fetch userID by scraping a user page...")

	// Make a request to a page known to contain window.ancestry.userId
	// The user's own facts page is a good candidate, but can be slow.
	// A simpler page might be /myancestry or any page that loads the global Ancestry object.
	// Let's try /myancestry as it should be light.
	endpoint := fmt.Sprintf("%s/myancestry", c.baseURL)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request for userID retrieval: %w", err)
	}
	// Mimic a browser request
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch page for userID retrieval: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.log.Printf("Error closing response body during userID retrieval: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to fetch page for userID retrieval, status %d: %s", resp.StatusCode, string(body))
	}

	html, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body for userID retrieval: %w", err)
	}

	// Scrape for window.ancestry.userId
	// Look for: userId:{value:"<THE_ID>",writable:!1}
	userIDRegex := regexp.MustCompile(`userId:{value:"(.*?)"`)
	matches := userIDRegex.FindStringSubmatch(string(html))
	if len(matches) > 1 {
		c.userID = matches[1]
		c.log.Printf("Successfully scraped userID: %s\n", c.userID)
		return c.userID, nil
	}

	return "", fmt.Errorf("failed to scrape userID from %s", endpoint)
}

// Close closes the API client and any associated resources
func (c *APIClient) Close() error {
	if c.loggingTransport != nil {
		return c.loggingTransport.Close()
	}
	return nil
}
