package ancestry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

const (
	// AncestryBaseURL is the main Ancestry.com URL
	AncestryBaseURL = "https://www.ancestry.com"

	// Network capture constants
	maxResponseBodySize = 1024 * 1024 // 1MB

	// Timing constants
	pageLoadDelay       = 2 * time.Second
	cloudflareMaxWait   = 30 * time.Second
	cloudflarePollDelay = 1 * time.Second
)

// Client represents a browser automation client for Ancestry.com
type Client struct {
	browser         *rod.Browser
	page            *rod.Page
	capturedRequest []*CapturedRequest
	captureEnabled  bool
	captureMutex    sync.Mutex // Protects capturedRequest and captureEnabled
}

// CapturedRequest represents a captured network request
type CapturedRequest struct {
	Method       string
	URL          string
	RequestBody  string
	ResponseBody string
	StatusCode   int
	ContentType  string
	Timestamp    time.Time
}

// NewClient creates a new Ancestry client with browser automation
func NewClient() (*Client, error) {
	// Try to find Chrome/Chromium binary
	path, found := launcher.LookPath()
	if !found {
		return nil, fmt.Errorf("Chrome/Chromium not found. Please install Chrome or Chromium browser")
	}

	fmt.Printf("Found browser at: %s\n", path)

	// Configure launcher with stealth flags to avoid bot detection
	l := launcher.New().
		Bin(path).
		Headless(false).
		Devtools(false)

	// Add Chrome flags as separate arguments
	l = l.Set("disable-blink-features", "AutomationControlled")
	l = l.Set("exclude-switches", "enable-automation")
	l = l.Set("disable-dev-shm-usage")
	l = l.Set("no-sandbox")
	l = l.Set("disable-setuid-sandbox")
	l = l.Set("window-size", "1920,1080")

	// Launch the browser
	fmt.Println("Launching browser...")
	u, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}

	fmt.Printf("Browser launched on: %s\n", u)

	// Connect to the browser
	browser := rod.New().ControlURL(u)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to browser: %w", err)
	}

	fmt.Println("Connected to browser successfully")

	return &Client{
		browser: browser,
	}, nil
}

// NavigateToAncestry opens Ancestry.com in the browser
func (c *Client) NavigateToAncestry() error {
	// Create a new page
	page := c.browser.MustPage()
	c.page = page

	// Inject comprehensive stealth scripts to hide automation markers
	stealthScript := `
		// Hide webdriver property
		Object.defineProperty(navigator, 'webdriver', {
			get: () => undefined
		});

		// Override chrome property with realistic values
		window.chrome = {
			runtime: {},
			loadTimes: function() {},
			csi: function() {},
			app: {}
		};

		// Override permissions
		const originalQuery = window.navigator.permissions.query;
		window.navigator.permissions.query = (parameters) => (
			parameters.name === 'notifications' ?
				Promise.resolve({ state: Notification.permission }) :
				originalQuery(parameters)
		);

		// Add realistic plugins
		Object.defineProperty(navigator, 'plugins', {
			get: () => {
				return [
					{
						0: {type: "application/x-google-chrome-pdf", suffixes: "pdf", description: "Portable Document Format"},
						description: "Portable Document Format",
						filename: "internal-pdf-viewer",
						length: 1,
						name: "Chrome PDF Plugin"
					},
					{
						0: {type: "application/pdf", suffixes: "pdf", description: "Portable Document Format"},
						description: "Portable Document Format",
						filename: "mhjfbmdgcfjbbpaeojofohoefgiehjai",
						length: 1,
						name: "Chrome PDF Viewer"
					},
					{
						0: {type: "application/x-nacl", suffixes: "", description: "Native Client Executable"},
						1: {type: "application/x-pnacl", suffixes: "", description: "Portable Native Client Executable"},
						description: "",
						filename: "internal-nacl-plugin",
						length: 2,
						name: "Native Client"
					}
				];
			}
		});

		// Override languages
		Object.defineProperty(navigator, 'languages', {
			get: () => ['en-US', 'en']
		});

		// Override hardwareConcurrency
		Object.defineProperty(navigator, 'hardwareConcurrency', {
			get: () => 8
		});

		// Override deviceMemory
		Object.defineProperty(navigator, 'deviceMemory', {
			get: () => 8
		});

		// Override platform
		Object.defineProperty(navigator, 'platform', {
			get: () => 'MacIntel'
		});

		// Override vendor
		Object.defineProperty(navigator, 'vendor', {
			get: () => 'Google Inc.'
		});

		// Override maxTouchPoints
		Object.defineProperty(navigator, 'maxTouchPoints', {
			get: () => 0
		});

		// Mock battery API
		if (!navigator.getBattery) {
			navigator.getBattery = () => Promise.resolve({
				charging: true,
				chargingTime: 0,
				dischargingTime: Infinity,
				level: 1
			});
		}

		// Mock media devices
		if (navigator.mediaDevices && navigator.mediaDevices.enumerateDevices) {
			const originalEnumerateDevices = navigator.mediaDevices.enumerateDevices;
			navigator.mediaDevices.enumerateDevices = () => {
				return originalEnumerateDevices().then(devices => {
					return devices.length > 0 ? devices : [
						{deviceId: "default", kind: "audioinput", label: "", groupId: "default"},
						{deviceId: "default", kind: "audiooutput", label: "", groupId: "default"},
						{deviceId: "default", kind: "videoinput", label: "", groupId: "default"}
					];
				});
			};
		}

		// Canvas fingerprinting protection
		const originalToDataURL = HTMLCanvasElement.prototype.toDataURL;
		HTMLCanvasElement.prototype.toDataURL = function(type) {
			if (type === 'image/png' && this.width === 280 && this.height === 60) {
				// Likely fingerprinting attempt, add noise
				const context = this.getContext('2d');
				const imageData = context.getImageData(0, 0, this.width, this.height);
				for (let i = 0; i < imageData.data.length; i += 4) {
					imageData.data[i] = imageData.data[i] ^ Math.floor(Math.random() * 3);
				}
				context.putImageData(imageData, 0, 0);
			}
			return originalToDataURL.apply(this, arguments);
		};

		// WebGL fingerprinting protection
		const getParameter = WebGLRenderingContext.prototype.getParameter;
		WebGLRenderingContext.prototype.getParameter = function(parameter) {
			// Randomize WebGL vendor/renderer
			if (parameter === 37445) {
				return 'Intel Inc.';
			}
			if (parameter === 37446) {
				return 'Intel Iris OpenGL Engine';
			}
			return getParameter.apply(this, arguments);
		};

		// AudioContext fingerprinting protection
		const AudioContext = window.AudioContext || window.webkitAudioContext;
		if (AudioContext) {
			const originalCreateAnalyser = AudioContext.prototype.createAnalyser;
			AudioContext.prototype.createAnalyser = function() {
				const analyser = originalCreateAnalyser.apply(this, arguments);
				const originalGetFloatFrequencyData = analyser.getFloatFrequencyData;
				analyser.getFloatFrequencyData = function(array) {
					originalGetFloatFrequencyData.apply(this, arguments);
					// Add slight noise to prevent fingerprinting
					for (let i = 0; i < array.length; i++) {
						array[i] += Math.random() * 0.0001;
					}
				};
				return analyser;
			};
		}

		// Connection API
		if (navigator.connection) {
			Object.defineProperties(navigator.connection, {
				downlink: { get: () => 10 },
				effectiveType: { get: () => '4g' },
				rtt: { get: () => 50 },
				saveData: { get: () => false }
			});
		}

		// Screen properties
		Object.defineProperty(screen, 'availWidth', {
			get: () => 1920
		});
		Object.defineProperty(screen, 'availHeight', {
			get: () => 1080
		});
		Object.defineProperty(screen, 'width', {
			get: () => 1920
		});
		Object.defineProperty(screen, 'height', {
			get: () => 1080
		});
		Object.defineProperty(screen, 'colorDepth', {
			get: () => 24
		});
		Object.defineProperty(screen, 'pixelDepth', {
			get: () => 24
		});

		// Timezone consistency
		Date.prototype.getTimezoneOffset = function() {
			return 300; // EST/CDT
		};

		// DevTools detection evasion
		// Block common DevTools detection methods
		const devtools = /./;
		devtools.toString = function() {
			return 'function RegExp() { [native code] }';
		};

		// Prevent detection via console properties
		const originalLog = console.log;
		Object.defineProperty(console, '_commandLineAPI', {
			get: () => undefined
		});

		// Block detection via debugger statement timing
		let checkCount = 0;
		const originalDate = Date;
		Date = class extends originalDate {
			constructor(...args) {
				if (args.length === 0) {
					super();
					checkCount++;
					if (checkCount > 100) {
						checkCount = 0;
					}
				} else {
					super(...args);
				}
			}
		};

		// Prevent outerHeight/outerWidth detection
		const originalOuterHeight = window.outerHeight;
		const originalOuterWidth = window.outerWidth;
		Object.defineProperty(window, 'outerHeight', {
			get: () => window.innerHeight
		});
		Object.defineProperty(window, 'outerWidth', {
			get: () => window.innerWidth
		});

		// Block Firebug detection
		window.firebug = undefined;

		// Block Chrome DevTools detection via toString
		const originalToString = Function.prototype.toString;
		Function.prototype.toString = function() {
			if (this === window.alert || this === window.prompt || this === window.confirm) {
				return 'function ' + this.name + '() { [native code] }';
			}
			return originalToString.call(this);
		};

		// Prevent detection via element.id getter
		const elementProto = Element.prototype;
		const originalGetAttribute = elementProto.getAttribute;
		elementProto.getAttribute = function(name) {
			if (name === 'id' && this.tagName === 'IFRAME') {
				return null;
			}
			return originalGetAttribute.call(this, name);
		};
	`

	// Execute stealth script before navigation
	if _, err := c.page.Eval(stealthScript); err != nil {
		fmt.Printf("Warning: Failed to inject stealth script: %v\n", err)
		fmt.Println("Continuing anyway, but bot detection may be more likely...")
	}

	// Navigate to Ancestry.com
	if err := c.page.Navigate(AncestryBaseURL); err != nil {
		return fmt.Errorf("failed to navigate to Ancestry.com: %w", err)
	}

	// Wait for the page to load
	if err := c.page.WaitLoad(); err != nil {
		return fmt.Errorf("failed to wait for page load: %w", err)
	}

	// Add a small delay to ensure page is fully rendered
	time.Sleep(pageLoadDelay)

	// Check if Cloudflare challenge is present
	if err := c.handleCloudflareChallenge(); err != nil {
		return fmt.Errorf("cloudflare challenge handling failed: %w", err)
	}

	return nil
}

// handleCloudflareChallenge detects and waits for Cloudflare challenge completion
func (c *Client) handleCloudflareChallenge() error {
	// Check if title element exists
	has, titleElem, err := c.page.Has("title")
	if err != nil || !has {
		// No title element, not a Cloudflare challenge
		return nil
	}

	// Get the title text
	title, err := titleElem.Text()
	if err != nil {
		// Can't read title, assume no challenge
		return nil
	}

	// Check if it's a Cloudflare challenge
	if !strings.Contains(title, "Just a moment") && !strings.Contains(title, "Cloudflare") {
		// Not a Cloudflare challenge
		return nil
	}

	fmt.Println("\n⚠️  Cloudflare challenge detected!")
	fmt.Println("Please complete the challenge in the browser window...")
	fmt.Printf("Waiting up to %v for you to complete it...\n", cloudflareMaxWait)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cloudflareMaxWait)
	defer cancel()

	// Poll for challenge completion
	ticker := time.NewTicker(cloudflarePollDelay)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("cloudflare challenge not completed in time (%v)", cloudflareMaxWait)
		case <-ticker.C:
			// Check if title has changed
			has, titleElem, err := c.page.Has("title")
			if err != nil || !has {
				// Title element disappeared or error, assume challenge passed
				fmt.Println("✓ Challenge completed!")
				return nil
			}

			currentTitle, err := titleElem.Text()
			if err != nil {
				// Can't read title, continue waiting
				continue
			}

			if !strings.Contains(currentTitle, "Just a moment") && !strings.Contains(currentTitle, "Cloudflare") {
				fmt.Println("✓ Challenge completed!")
				return nil
			}
		}
	}
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

// EnableNetworkCapture enables capturing of network requests
func (c *Client) EnableNetworkCapture() error {
	if c.page == nil {
		return fmt.Errorf("no page available")
	}

	c.captureMutex.Lock()
	c.captureEnabled = true
	c.capturedRequest = make([]*CapturedRequest, 0)
	c.captureMutex.Unlock()

	// Enable network events
	router := c.page.HijackRequests()
	router.MustAdd("*", func(ctx *rod.Hijack) {
		// Load the request
		ctx.MustLoadResponse()

		// Filter for API requests (JSON responses, XHR/Fetch)
		contentType := ctx.Response.Headers().Get("Content-Type")
		requestURL := ctx.Request.URL().String()
		isAPI := strings.Contains(contentType, "application/json") ||
			strings.Contains(contentType, "application/javascript") ||
			strings.Contains(requestURL, "/api/") ||
			strings.Contains(requestURL, "/cgi-bin/")

		// Check if capture is enabled (thread-safe read)
		c.captureMutex.Lock()
		enabled := c.captureEnabled
		c.captureMutex.Unlock()

		if isAPI && enabled {
			captured := &CapturedRequest{
				Method:      ctx.Request.Method(),
				URL:         ctx.Request.URL().String(),
				StatusCode:  ctx.Response.Payload().ResponseCode,
				ContentType: contentType,
				Timestamp:   time.Now(),
			}

			// Capture request body if present
			if ctx.Request.Body() != "" {
				captured.RequestBody = ctx.Request.Body()
			}

			// Capture response body (limit size to avoid huge responses)
			body := ctx.Response.Body()
			if len(body) > 0 && len(body) < maxResponseBodySize {
				captured.ResponseBody = body
			} else if len(body) >= maxResponseBodySize {
				fmt.Printf("   [Warning] Response body too large (%d bytes), truncated\n", len(body))
			}

			// Thread-safe append
			c.captureMutex.Lock()
			c.capturedRequest = append(c.capturedRequest, captured)
			c.captureMutex.Unlock()
		}
	})

	go router.Run()

	return nil
}

// DisableNetworkCapture stops capturing network requests
func (c *Client) DisableNetworkCapture() {
	c.captureMutex.Lock()
	defer c.captureMutex.Unlock()
	c.captureEnabled = false
}

// GetCapturedRequests returns all captured network requests (returns a copy to prevent race conditions)
func (c *Client) GetCapturedRequests() []*CapturedRequest {
	c.captureMutex.Lock()
	defer c.captureMutex.Unlock()

	// Return a copy to prevent concurrent access issues
	result := make([]*CapturedRequest, len(c.capturedRequest))
	copy(result, c.capturedRequest)
	return result
}

// ClearCapturedRequests clears the captured requests list
func (c *Client) ClearCapturedRequests() {
	c.captureMutex.Lock()
	defer c.captureMutex.Unlock()
	c.capturedRequest = make([]*CapturedRequest, 0)
}
