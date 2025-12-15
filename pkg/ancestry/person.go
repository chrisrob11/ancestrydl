package ancestry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

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

// GetPersonFactsFromHTML scrapes the "Facts" page for the researchData JSON
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
