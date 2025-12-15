package ancestry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// Client handles browser automation using Rod
type Client struct {
	browser          *rod.Browser
	page             *rod.Page
	capturedRequests []*CapturedRequest
	mu               sync.Mutex
}

// NewClient creates a new Client with a headful browser
func NewClient() (*Client, error) {
	// Launch headful browser so user can see/interact if needed (e.g. for CAPTCHA)
	u := launcher.New().Headless(false).MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect()

	// Create a new page
	page := browser.MustPage()

	return &Client{
		browser: browser,
		page:    page,
	}, nil
}

// Close closes the browser
func (c *Client) Close() error {
	if c.browser != nil {
		return c.browser.Close()
	}
	return nil
}

// NavigateToAncestry navigates to the Ancestry homepage
func (c *Client) NavigateToAncestry() error {
	if c.page == nil {
		return fmt.Errorf("page is nil")
	}
	c.page.MustNavigate("https://www.ancestry.com")
	return c.page.WaitLoad()
}

// GetAncestrySessionCookies extracts cookies for ancestry.com domains
func (c *Client) GetAncestrySessionCookies() ([]*proto.NetworkCookie, error) {
	if c.page == nil {
		return nil, fmt.Errorf("page is nil")
	}
	// Get all cookies for the current URL
	cookies, err := c.page.Cookies([]string{"https://www.ancestry.com"})
	if err != nil {
		return nil, err
	}
	return cookies, nil
}

// GetPage returns the current page
func (c *Client) GetPage() *rod.Page {
	return c.page
}

// GetCapturedRequests returns the captured requests
func (c *Client) GetCapturedRequests() []*CapturedRequest {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Return a copy to avoid race conditions if called while capturing
	requests := make([]*CapturedRequest, len(c.capturedRequests))
	copy(requests, c.capturedRequests)
	return requests
}

// EnableNetworkCapture starts capturing network traffic
func (c *Client) EnableNetworkCapture() error {
	if c.page == nil {
		return fmt.Errorf("page is nil")
	}

	// Enable network domain
	// Note: go-rod enables Network by default for some things, but explicit enablement is good.
	// However, usually we just listen to events.

	// Map to store requests by ID to match with responses
	pendingRequests := make(map[string]*CapturedRequest)
	var mapMu sync.Mutex

	// Listen for requests
	go c.page.EachEvent(func(e *proto.NetworkRequestWillBeSent) {
		req := &CapturedRequest{
			URL:       e.Request.URL,
			Method:    e.Request.Method,
			Timestamp: time.Now(), // e.Timestamp is monotonic, not wall clock
		}

		if e.Request.HasPostData {
			req.RequestBody = e.Request.PostData
		}

		mapMu.Lock()
		pendingRequests[string(e.RequestID)] = req
		mapMu.Unlock()
	})

	// Listen for responses
	go c.page.EachEvent(func(e *proto.NetworkResponseReceived) {
		mapMu.Lock()
		req, ok := pendingRequests[string(e.RequestID)]
		if !ok {
			mapMu.Unlock()
			return
		}
		// Remove from pending map as we processed it (mostly - body comes later usually)
		// But let's keep it simple and just update it.
		// Actually, we need to wait for loading finished to get body, but getting body requires separate command.
		// For now, let's just capture status and headers.
		delete(pendingRequests, string(e.RequestID))
		mapMu.Unlock()

		req.StatusCode = e.Response.Status
		req.ContentType = e.Response.MIMEType

		// Add to captured requests
		c.mu.Lock()
		c.capturedRequests = append(c.capturedRequests, req)
		c.mu.Unlock()
	})

	return nil
}

// SerializeCookies converts cookies to JSON string
func SerializeCookies(cookies []*proto.NetworkCookie) (string, error) {
	data, err := json.Marshal(cookies)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DeserializeCookies converts JSON string to cookies
func DeserializeCookies(data string) ([]*proto.NetworkCookie, error) {
	var cookies []*proto.NetworkCookie
	if err := json.Unmarshal([]byte(data), &cookies); err != nil {
		return nil, err
	}
	return cookies, nil
}

// CookiesToHTTPCookies converts rod cookies to http.Cookie
func CookiesToHTTPCookies(rodCookies []*proto.NetworkCookie) []*http.Cookie {
	var httpCookies []*http.Cookie
	for _, rc := range rodCookies {
		c := &http.Cookie{
			Name:   rc.Name,
			Value:  rc.Value,
			Domain: rc.Domain,
			Path:   rc.Path,
			Secure: rc.Secure,
		}
		// Expires is float64 seconds since epoch
		if rc.Expires > 0 {
			c.Expires = time.Unix(int64(rc.Expires), 0)
		}
		httpCookies = append(httpCookies, c)
	}
	return httpCookies
}
