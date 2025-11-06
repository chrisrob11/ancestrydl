package ancestry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/go-rod/rod/lib/proto"
	"golang.org/x/net/publicsuffix"
)

// APIClient handles HTTP requests to Ancestry.com APIs
type APIClient struct {
	httpClient       *http.Client
	baseURL          string
	loggingTransport *loggingTransport // For verbose mode
}

// Tree represents an Ancestry family tree
// This struct supports both the old and new tree list APIs
type Tree struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"desc,omitempty"`
	Owner             string    `json:"owner,omitempty"`        // From /api/media/viewer/api/trees/list
	OwnerUserID       string    `json:"ownerUserId,omitempty"`  // From old API
	DateModified      time.Time `json:"dateModified,omitempty"` // Old format
	DateCreated       time.Time `json:"dateCreated,omitempty"`  // Old format
	CreatedOn         time.Time `json:"createdOn,omitempty"`    // New format
	ModifiedOn        time.Time `json:"modifiedOn,omitempty"`   // New format
	CanSeeLiving      bool      `json:"canSeeLiving,omitempty"` // From /api/media/viewer/api/trees/list
	TotalInvitedCount int       `json:"totalInvitedCount,omitempty"`
	SH                bool      `json:"sh,omitempty"`
}

// TreeListResponse represents the response from the tree list API
type TreeListResponse struct {
	Trees []Tree `json:"trees"`
	Count int    `json:"count"`
}

// Person represents a person in the family tree (matches Ancestry API structure)
type Person struct {
	GID           map[string]interface{} `json:"gid"`
	Names         []Name                 `json:"Names,omitempty"`
	Genders       []Gender               `json:"Genders,omitempty"`
	Events        []Event                `json:"Events,omitempty"`
	L             interface{}            `json:"l,omitempty"`
	Lus           interface{}            `json:"lus,omitempty"`
	MD            string                 `json:"md,omitempty"` // Modified date
	CD            string                 `json:"cd,omitempty"` // Created date
	Kinships      []interface{}          `json:"Kinships,omitempty"`
	KinshipLabel  string                 `json:"kinshipLabel,omitempty"`
	Family        []FamilyMember         `json:"Family,omitempty"` // Family relationships
	PID           string                 `json:"pid,omitempty"`
	IsLiving      bool                   `json:"isLiving,omitempty"`
	GivenName     string                 `json:"gname,omitempty"` // Flat field for given name
	Surname       string                 `json:"sname,omitempty"` // Flat field for surname
	Gender        string                 `json:"gender,omitempty"`
	EventsSummary []interface{}          `json:"events,omitempty"`
}

// FamilyMember represents a family relationship
type FamilyMember struct {
	Type string                 `json:"t"`    // F=Father, M=Mother, H=Husband, W=Wife, C=Child
	TGID map[string]interface{} `json:"tgid"` // Target person ID
	CD   string                 `json:"cd,omitempty"`
	MD   string                 `json:"md,omitempty"`
	Mod  string                 `json:"mod,omitempty"`
}

// GetPersonID extracts the person ID from the GID map
// The person ID is stored in gid.v (e.g., {"v":"232573524428:1030:197283789"})
// Returns the full ID including tree context
func (p *Person) GetPersonID() string {
	// First check if PID is already set (for backwards compatibility)
	if p.PID != "" {
		return p.PID
	}

	// Extract from GID map
	if p.GID != nil {
		if v, ok := p.GID["v"].(string); ok {
			return v
		}
	}

	return ""
}

// GetShortPersonID extracts just the person identifier without tree context
// For example, "232573524428:1030:197283789" becomes "232573524428"
func (p *Person) GetShortPersonID() string {
	fullID := p.GetPersonID()
	if fullID == "" {
		return ""
	}

	// Split on colon and return first part
	parts := strings.Split(fullID, ":")
	return parts[0]
}

// GetDisplayName returns the person's name from the Names array or falls back to flat fields
func (p *Person) GetDisplayName() string {
	// Try to get name from Names array first
	if len(p.Names) > 0 {
		givenName := p.Names[0].GivenName
		surname := p.Names[0].Surname
		if givenName != "" || surname != "" {
			return strings.TrimSpace(fmt.Sprintf("%s %s", givenName, surname))
		}
	}

	// Fall back to flat fields
	if p.GivenName != "" || p.Surname != "" {
		return strings.TrimSpace(fmt.Sprintf("%s %s", p.GivenName, p.Surname))
	}

	return ""
}

// Name represents a person's name in structured format
type Name struct {
	ID        string `json:"id"`
	GivenName string `json:"g"` // Given name / first name
	Surname   string `json:"s"` // Surname / last name
}

// Gender represents a person's gender
type Gender struct {
	ID     string `json:"id"`
	Gender string `json:"g"` // "m" or "f"
}

// Event represents a life event (birth, death, marriage, etc.)
type Event struct {
	ID          string                   `json:"id,omitempty"`
	Type        string                   `json:"t,omitempty"`
	Date        interface{}              `json:"d,omitempty"`
	NPS         []map[string]interface{} `json:"nps,omitempty"`  // Nested place structure
	Description string                   `json:"desc,omitempty"` // Event description/notes
}

// FamilyViewResponse represents the response from the newfamilyview API
type FamilyViewResponse struct {
	V       string      `json:"v"` // Version string like "3.0"
	Persons []Person    `json:"Persons"`
	Focus   interface{} `json:"focus,omitempty"` // Can be object or string
}

// TreeInfo represents tree metadata
type TreeInfo struct {
	TreeID          string `json:"treeId"`
	TreeName        string `json:"treeName"`
	TreeDescription string `json:"treeDescription,omitempty"`
	IsPrivate       bool   `json:"isPrivate"`
	PersonCount     int    `json:"personCount,omitempty"`
}

// FocusHistoryResponse represents the focus history with person data
type FocusHistoryResponse struct {
	History []FocusHistoryItem `json:"History"`
	Persons map[string]Person  `json:"Persons"`
}

// FocusHistoryItem represents a single item in focus history
type FocusHistoryItem struct {
	PID       string `json:"pid"`
	Timestamp int64  `json:"ts"`
}

// UserData represents user account information
type UserData struct {
	User                     map[string]interface{} `json:"user"`
	HasHints                 bool                   `json:"hasHints"`
	MostRecentlyViewedTreeID string                 `json:"mostRecentlyViewedTreeId"`
	NotificationsCount       int                    `json:"notificationsCount"`
	HintCount                int                    `json:"hintcount"`
}

// PersonMedia represents media attached to a person
type PersonMedia struct {
	PersonID   string      `json:"personId,omitempty"`
	TreeID     string      `json:"treeId,omitempty"`
	MediaItems []MediaItem `json:"mediaItems,omitempty"`
}

// MediaItem represents a single media item (photo, document, etc.)
type MediaItem struct {
	MediaID     string                 `json:"mediaId,omitempty"`
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type,omitempty"`
	URL         string                 `json:"url,omitempty"`
	ThumbURL    string                 `json:"thumbUrl,omitempty"`
	Category    string                 `json:"category,omitempty"`
	Raw         map[string]interface{} `json:"-"` // Store raw response for debugging
}

// InitialState represents the window.INITIAL_STATE object in person pages
type InitialState struct {
	Redux ReduxState `json:"redux"`
}

// ReduxState is part of the INITIAL_STATE
type ReduxState struct {
	Person PersonState `json:"person"`
}

// PersonState holds person-specific page data
type PersonState struct {
	PageData PageData `json:"pageData"`
}

// PageData contains the person facts
type PageData struct {
	PersonFacts map[string]interface{} `json:"personFacts"`
}

// PersonFact represents a single fact, which may have media
type PersonFact struct {
	PrimaryMediaItem *PrimaryMediaItem `json:"primaryMediaItem"`
}

// PrimaryMediaItem contains direct URLs to media with metadata
type PrimaryMediaItem struct {
	Type        string `json:"type"`
	URL         string `json:"url"`
	PreviewURL  string `json:"previewUrl"`
	MediaID     string `json:"mediaId"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Subcategory string `json:"subcategory"`
	Description string `json:"description"`
	Date        string `json:"date"`
}

// MediaViewerResponse represents the response from /api/media/viewer/v1/trees/{treeId}/people/{personId}
type MediaViewerResponse struct {
	MediaCount int                 `json:"mediaCount"`
	HasMedia   bool                `json:"hasMedia"`
	Objects    []MediaViewerObject `json:"objects"`
}

// MediaViewerObject represents a single media item from the media viewer API
type MediaViewerObject struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Type         string `json:"type"`
	Category     string `json:"category"`
	Subcategory  string `json:"subcategory"`
	URL          string `json:"url"`
	CollectionID int    `json:"collectionId"`
	Description  string `json:"description"`
	Date         string `json:"date"`
	PreviewURL   string `json:"previewUrl"`
}

// ResearchData represents the window.researchData object embedded in Facts pages
type ResearchData struct {
	PersonFacts []PersonFactDetail `json:"PersonFacts"`
}

// PersonFactDetail represents a single fact/event with complete details
type PersonFactDetail struct {
	Type        int                    `json:"Type"`
	TypeString  string                 `json:"TypeString"`
	Title       string                 `json:"Title,omitempty"` // Custom event title (like "Prison")
	Place       string                 `json:"Place,omitempty"`
	PlaceGpids  map[string]interface{} `json:"PlaceGpids,omitempty"`
	Description string                 `json:"Description,omitempty"`
	Date        interface{}            `json:"Date,omitempty"`
}

// NewAPIClient creates a new API client with the given cookies
func NewAPIClient(cookies []*proto.NetworkCookie, verbose bool) (*APIClient, error) {
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

// Close closes the API client and any associated resources
func (c *APIClient) Close() error {
	if c.loggingTransport != nil {
		return c.loggingTransport.Close()
	}
	return nil
}

// ListTrees retrieves all trees (owned and shared) for the authenticated user
func (c *APIClient) ListTrees() ([]Tree, error) {
	// Use the media viewer API which returns ALL trees including shared ones
	endpoint := fmt.Sprintf("%s/api/media/viewer/api/trees/list", c.baseURL)
	reqURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add timestamp parameter
	query := reqURL.Query()
	query.Set("timestamp", fmt.Sprintf("%d", time.Now().UnixMilli()))
	reqURL.RawQuery = query.Encode()

	// Create request
	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Referer", "https://www.ancestry.com/")

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// This endpoint returns an array directly, not wrapped in {trees: [...]}
	var trees []Tree
	if err := json.NewDecoder(resp.Body).Decode(&trees); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return trees, nil
}

// GetTreeInfo retrieves metadata about a specific tree
func (c *APIClient) GetTreeInfo(treeID string) (*TreeInfo, error) {
	endpoint := fmt.Sprintf("%s/api/treeviewer/tree/%s/info", c.baseURL, treeID)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Referer", "https://www.ancestry.com/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var treeInfo TreeInfo
	if err := json.NewDecoder(resp.Body).Decode(&treeInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &treeInfo, nil
}

// GetFamilyView retrieves comprehensive tree data for multiple generations
// This is the primary endpoint for downloading tree information
func (c *APIClient) GetFamilyView(treeID, focusPersonID string, genUp, genDown int) (*FamilyViewResponse, error) {
	endpoint := fmt.Sprintf("%s/api/treeviewer/tree/newfamilyview/%s", c.baseURL, treeID)

	reqURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	query := reqURL.Query()
	query.Set("focusPersonId", focusPersonID)
	query.Set("isFocus", "true")
	query.Set("view", "family")
	query.Set("genup", fmt.Sprintf("%d", genUp))
	query.Set("gendown", fmt.Sprintf("%d", genDown))
	query.Set("ts", fmt.Sprintf("%d", time.Now().UnixMilli()))
	reqURL.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Referer", "https://www.ancestry.com/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var familyView FamilyViewResponse
	if err := json.NewDecoder(resp.Body).Decode(&familyView); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &familyView, nil
}

// GetRootPerson retrieves the root person of a tree
func (c *APIClient) GetRootPerson(treeID string) (*Person, error) {
	endpoint := fmt.Sprintf("%s/api/treesui-list/trees/%s/rootperson", c.baseURL, treeID)

	reqURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	query := reqURL.Query()
	query.Set("expires", fmt.Sprintf("%d", time.Now().UnixMilli()))
	query.Set("isGetFullPersonObject", "true")
	reqURL.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Referer", "https://www.ancestry.com/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var person Person
	if err := json.NewDecoder(resp.Body).Decode(&person); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &person, nil
}

// GetFocusHistory retrieves the navigation history with person data
func (c *APIClient) GetFocusHistory(treeID string) (*FocusHistoryResponse, error) {
	endpoint := fmt.Sprintf("%s/api/treeviewer/getFocusHistory", c.baseURL)

	reqURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	query := reqURL.Query()
	query.Set("tid", treeID)
	query.Set("ts", fmt.Sprintf("%d", time.Now().UnixMilli()))
	reqURL.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Referer", "https://www.ancestry.com/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var history FocusHistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &history, nil
}

// GetComments retrieves comments for multiple persons in a tree
func (c *APIClient) GetComments(treeID string, personIDs []string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("%s/api/treeviewer/comments/tree/%s", c.baseURL, treeID)

	reqURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	query := reqURL.Query()
	for _, pid := range personIDs {
		query.Add("pid", pid)
	}
	reqURL.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Referer", "https://www.ancestry.com/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var comments map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return comments, nil
}

// GetMediaImage downloads an image from Ancestry media storage
func (c *APIClient) GetMediaImage(namespace, mediaGUID string, maxWidth, maxHeight int) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/api/media/retrieval/v2/image/namespaces/%s/media/%s.jpg",
		c.baseURL, namespace, mediaGUID)

	reqURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	query := reqURL.Query()
	query.Set("client", "Ancestry.Trees")
	if maxWidth > 0 {
		query.Set("maxWidth", fmt.Sprintf("%d", maxWidth))
	}
	if maxHeight > 0 {
		query.Set("maxHeight", fmt.Sprintf("%d", maxHeight))
	}
	reqURL.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	return imageData, nil
}

// GetUserData retrieves user account information
func (c *APIClient) GetUserData() (*UserData, error) {
	endpoint := fmt.Sprintf("%s/api/navheaderdata/v1/header/data/user", c.baseURL)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Referer", "https://www.ancestry.com/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var userData UserData
	if err := json.NewDecoder(resp.Body).Decode(&userData); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &userData, nil
}

// GetAllPersons retrieves all persons in a tree with pagination support
// Returns persons sorted by surname, given name, and ID
func (c *APIClient) GetAllPersons(treeID string, page, limit int) ([]Person, error) {
	endpoint := fmt.Sprintf("%s/api/treesui-list/trees/%s/persons", c.baseURL, treeID)

	reqURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	query := reqURL.Query()
	query.Set("expires", fmt.Sprintf("%d", time.Now().UnixMilli()))
	query.Set("fn", "")
	query.Set("ln", "")
	query.Set("name", "")
	query.Set("tags", "")
	query.Set("sort", "sname,gname,id")
	query.Set("page", fmt.Sprintf("%d", page))
	query.Set("limit", fmt.Sprintf("%d", limit))
	query.Set("fields", "NAMES,EVENTS")
	reqURL.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Referer", fmt.Sprintf("https://www.ancestry.com/family-tree/tree/%s/listofallpeople", treeID))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var persons []Person
	if err := json.NewDecoder(resp.Body).Decode(&persons); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return persons, nil
}

// GetPersonsCount retrieves the total count of persons in a tree
func (c *APIClient) GetPersonsCount(treeID string) (int, error) {
	endpoint := fmt.Sprintf("%s/api/treesui-list/trees/%s/persons/count", c.baseURL, treeID)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Referer", fmt.Sprintf("https://www.ancestry.com/family-tree/tree/%s/listofallpeople", treeID))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var count int
	if err := json.NewDecoder(resp.Body).Decode(&count); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	return count, nil
}

// GetPersonMedia retrieves all media attached to a person
func (c *APIClient) GetPersonMedia(treeID, personID string) (*PersonMedia, error) {
	endpoint := fmt.Sprintf("%s/api/media/viewer/v1/trees/%s/people/%s", c.baseURL, treeID, personID)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", fmt.Sprintf("https://www.ancestry.com/family-tree/person/tree/%s/person/%s", treeID, personID))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Debug: Print response structure for first person
	// fmt.Printf("DEBUG: GetPersonMedia response keys: %v\n", reflect.ValueOf(result).MapKeys())
	// fmt.Printf("DEBUG: Full response: %s\n", string(bodyBytes))

	// Parse the media response into our struct
	personMedia := &PersonMedia{
		PersonID: personID,
		TreeID:   treeID,
	}

	extractPersonMedia(personMedia, result)

	return personMedia, nil
}

func extractPersonMedia(personMedia *PersonMedia, result map[string]interface{}) {
	// The API may return different structures, so we'll handle this flexibly
	if mediaArray, ok := result["media"].([]interface{}); ok {
		for _, item := range mediaArray {
			if mediaMap, ok := item.(map[string]interface{}); ok {
				mediaItem := MediaItem{
					Raw: mediaMap,
				}

				// Extract common fields
				if id, ok := mediaMap["id"].(string); ok {
					mediaItem.MediaID = id
				} else if mediaGUID, ok := mediaMap["mediaGuid"].(string); ok {
					mediaItem.MediaID = mediaGUID
				}

				if title, ok := mediaMap["title"].(string); ok {
					mediaItem.Title = title
				}

				if desc, ok := mediaMap["description"].(string); ok {
					mediaItem.Description = desc
				}

				if mediaType, ok := mediaMap["type"].(string); ok {
					mediaItem.Type = mediaType
				}

				if category, ok := mediaMap["category"].(string); ok {
					mediaItem.Category = category
				}

				if url, ok := mediaMap["url"].(string); ok {
					mediaItem.URL = url
				}

				if thumbURL, ok := mediaMap["thumbnailUrl"].(string); ok {
					mediaItem.ThumbURL = thumbURL
				}

				personMedia.MediaItems = append(personMedia.MediaItems, mediaItem)
			}
		}
	}
}

// GetPersonMediaFromAPI fetches media items using the media viewer API
func (c *APIClient) GetPersonMediaFromAPI(treeID, personID string) ([]PrimaryMediaItem, error) {
	// Extract just the person ID (first part before colon)
	shortPersonID := personID
	if parts := strings.Split(personID, ":"); len(parts) > 0 {
		shortPersonID = parts[0]
	}

	endpoint := fmt.Sprintf("%s/api/media/viewer/v1/trees/%s/people/%s", c.baseURL, treeID, shortPersonID)

	// Add query parameters
	reqURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	query := reqURL.Query()
	query.Set("excludeInlineStories", "1")
	query.Set("collectionId", "1030")
	query.Set("lcid", "1033")
	query.Set("page", "1")
	query.Set("rows", "100") // Get up to 100 media items
	query.Set("sort", "-created")
	reqURL.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch media: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("media API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var mediaResp MediaViewerResponse
	if err := json.NewDecoder(resp.Body).Decode(&mediaResp); err != nil {
		// If no media, return empty slice instead of error
		if strings.Contains(err.Error(), "cannot unmarshal") {
			return []PrimaryMediaItem{}, nil
		}
		return nil, fmt.Errorf("failed to decode media response: %w", err)
	}

	// Convert to PrimaryMediaItem format
	var mediaItems []PrimaryMediaItem
	for _, obj := range mediaResp.Objects {
		// Prepend base URL if the URL is relative
		mediaURL := obj.URL
		if strings.HasPrefix(mediaURL, "/") {
			mediaURL = c.baseURL + mediaURL
		}

		previewURL := obj.PreviewURL
		if strings.HasPrefix(previewURL, "/") {
			previewURL = c.baseURL + previewURL
		}

		mediaItems = append(mediaItems, PrimaryMediaItem{
			URL:         mediaURL,
			PreviewURL:  previewURL,
			Type:        obj.Type,
			Category:    obj.Category,
			Subcategory: obj.Subcategory,
			MediaID:     obj.ID,
			Title:       obj.Title,
			Description: obj.Description,
			Date:        obj.Date,
		})
	}

	return mediaItems, nil
}

// GetPersonFactsAndMedia scrapes the person's facts page to find media URLs (DEPRECATED - use GetPersonMediaFromAPI)
func (c *APIClient) GetPersonFactsAndMedia(treeID, personID string) ([]PrimaryMediaItem, error) {
	// Extract just the person ID (first part before colon)
	// Person IDs come as "232573524428:1030:197283789" but URLs only need "232573524428"
	shortPersonID := personID
	if parts := strings.Split(personID, ":"); len(parts) > 0 {
		shortPersonID = parts[0]
	}

	endpoint := fmt.Sprintf("%s/family-tree/person/tree/%s/person/%s/facts", c.baseURL, treeID, shortPersonID)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for facts page: %w", err)
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Referer", fmt.Sprintf("https://www.ancestry.com/family-tree/tree/%s/family/familyview", treeID))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch facts page: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("facts page request failed with status %d (URL: %s)", resp.StatusCode, endpoint)
	}

	html, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read facts page body: %w", err)
	}

	// Extract the INITIAL_STATE JSON from the HTML
	htmlContent := string(html)
	startMarker := "window.INITIAL_STATE = "
	startIndex := strings.Index(htmlContent, startMarker)
	if startIndex == -1 {
		fmt.Println("   [Debug] Could not find INITIAL_STATE in HTML content")
		return nil, nil // Return empty slice instead of error
	}

	// Find the end of the JSON object
	endMarker := `</script>`
	jsonStartIndex := startIndex + len(startMarker)
	jsonEndIndex := strings.Index(htmlContent[jsonStartIndex:], endMarker)
	if jsonEndIndex == -1 {
		return nil, fmt.Errorf("could not find end of INITIAL_STATE script tag")
	}
	jsonEndIndex += jsonStartIndex
	// The JSON blob is terminated by a semicolon
	jsonStr := strings.TrimRight(htmlContent[jsonStartIndex:jsonEndIndex], ";")

	var initialState InitialState
	if err := json.Unmarshal([]byte(jsonStr), &initialState); err != nil {
		return nil, fmt.Errorf("failed to unmarshal INITIAL_STATE JSON: %w", err)
	}

	// Extract media items
	mediaItems := extractMediaItems(initialState)

	fmt.Printf("   [Debug] Found %d media items for person %s\n", len(mediaItems), personID)
	return mediaItems, nil
}

func extractMediaItems(initialState InitialState) []PrimaryMediaItem {
	// Extract media items
	var mediaItems []PrimaryMediaItem
	if facts, ok := initialState.Redux.Person.PageData.PersonFacts["facts"].(map[string]interface{}); ok {
		if items, ok := facts["items"].([]interface{}); ok {
			for _, item := range items {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if media, ok := itemMap["primaryMediaItem"].(map[string]interface{}); ok {
						var mediaItem PrimaryMediaItem
						// Manual conversion from map to struct
						if mID, ok := media["mediaId"].(string); ok {
							mediaItem.MediaID = mID
						}
						if mType, ok := media["type"].(string); ok {
							mediaItem.Type = mType
						}
						if mURL, ok := media["url"].(string); ok {
							mediaItem.URL = mURL
						}
						if pURL, ok := media["previewUrl"].(string); ok {
							mediaItem.PreviewURL = pURL
						}

						if mediaItem.MediaID != "" && mediaItem.URL != "" {
							mediaItems = append(mediaItems, mediaItem)
						}
					}
				}
			}
		}
	}
	return mediaItems
}

// fetchFactsPageWithTimeout fetches the Facts page HTML with a specific timeout
func (c *APIClient) fetchFactsPageWithTimeout(endpoint, treeID string, timeout time.Duration) ([]byte, error) {
	// Create a new HTTP client with the specified timeout
	client := &http.Client{
		Jar:       c.httpClient.Jar,
		Timeout:   timeout,
		Transport: c.httpClient.Transport,
	}

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Referer", fmt.Sprintf("https://www.ancestry.com/family-tree/tree/%s/family/familyview", treeID))

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	html, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return html, nil
}

// GetPersonFactsFromHTML fetches the person's Facts page HTML and extracts window.researchData
// It will retry once on failure, waiting 30 seconds before the retry
func (c *APIClient) GetPersonFactsFromHTML(treeID, personID string) (*ResearchData, error) {
	// Extract just the person ID (first part before colon)
	shortPersonID := personID
	if parts := strings.Split(personID, ":"); len(parts) > 0 {
		shortPersonID = parts[0]
	}

	endpoint := fmt.Sprintf("%s/family-tree/person/tree/%s/person/%s/facts", c.baseURL, treeID, shortPersonID)

	// Try the request, with one retry on failure
	html, err := c.factsHTMLReq(endpoint, treeID)
	if err != nil {
		return nil, err
	}

	// Extract the window.researchData JSON from the HTML
	htmlContent := string(html)
	startMarker := "window.researchData = "
	startIndex := strings.Index(htmlContent, startMarker)
	if startIndex == -1 {
		return nil, nil // Return nil if no research data found (not an error)
	}

	// Find the end of the JSON object by counting braces
	jsonStartIndex := startIndex + len(startMarker)

	// Count braces to find the matching closing brace
	braceCount := 0
	inString := false
	escaped := false
	jsonEndIndex := jsonStartIndex

	for i := jsonStartIndex; i < len(htmlContent); i++ {
		ch := htmlContent[i]

		if escaped {
			escaped = false
			continue
		}

		if ch == '\\' {
			escaped = true
			continue
		}

		if ch == '"' {
			inString = !inString
			continue
		}

		if !inString {
			if ch == '{' {
				braceCount++
			} else if ch == '}' {
				braceCount--
				if braceCount == 0 {
					jsonEndIndex = i + 1
					break
				}
			}
		}
	}

	if braceCount != 0 {
		return nil, fmt.Errorf("could not find matching closing brace for window.researchData")
	}

	// Extract the JSON string
	jsonStr := htmlContent[jsonStartIndex:jsonEndIndex]

	// Parse the JSON (it's already valid JSON, not escaped)
	var researchData ResearchData
	if err := json.Unmarshal([]byte(jsonStr), &researchData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal window.researchData: %w", err)
	}

	return &researchData, nil
}

func (c *APIClient) factsHTMLReq(endpoint, treeID string) ([]byte, error) {
	// Try the request, with one retry on failure
	var html []byte
	var err error
	for attempt := 1; attempt <= 2; attempt++ {
		var timeout time.Duration
		if attempt == 1 {
			timeout = 30 * time.Second
		} else {
			// Second attempt: wait 30 seconds, then try with 35 second timeout
			time.Sleep(30 * time.Second)
			timeout = 35 * time.Second
		}

		html, err = c.fetchFactsPageWithTimeout(endpoint, treeID, timeout)
		if err == nil {
			break // Success!
		}

		// If this was the last attempt, return the error
		if attempt == 2 {
			return nil, fmt.Errorf("failed to fetch facts page after 2 attempts: %w", err)
		}
	}

	return html, nil
}

// DownloadFile downloads a file from a given URL
func (c *APIClient) DownloadFile(fileURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", fileURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}
	req.Header.Set("Accept", "image/webp,image/apng,image/*,*/*;q=0.8")
	req.Header.Set("Referer", c.baseURL+"/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d for URL %s", resp.StatusCode, fileURL)
	}

	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read downloaded file data: %w", err)
	}

	return fileData, nil
}
