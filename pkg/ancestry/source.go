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

// GetSource retrieves source record information by scraping the fact edit page
// and extracting the window.getFactEditData JSON object.
func (c *APIClient) GetSource(treeID, personID, sourceID, databaseID, recordID string) (*FactEditData, error) {
	reqURL, err := c.buildSourceURL(treeID, personID, sourceID, databaseID, recordID)
	if err != nil {
		return nil, err
	}

	shortPersonID := strings.Split(personID, ":")[0]
	var lastErr error

	// Implement retry mechanism for transient server errors
	for attempt := 1; attempt <= 3; attempt++ {
		factEditData, shouldRetry, err := c.performSourceAttempt(reqURL, treeID, shortPersonID, attempt)
		if err == nil {
			return factEditData, nil
		}
		lastErr = err
		if !shouldRetry {
			return nil, lastErr
		}
	}
	return nil, lastErr
}

func (c *APIClient) buildSourceURL(treeID, personID, sourceID, databaseID, recordID string) (*url.URL, error) {
	userID, err := c.GetUserID()
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticated user ID: %w", err)
	}
	if userID == "" {
		return nil, fmt.Errorf("authenticated user ID is empty, cannot construct source URL")
	}

	shortPersonID := strings.Split(personID, ":")[0]

	endpoint := fmt.Sprintf("%s/family-tree/person/factedit/user/%s/tree/%s/person/%s/source/%s",
		c.baseURL, userID, treeID, shortPersonID, sourceID)

	reqURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	query := reqURL.Query()
	if databaseID != "" {
		query.Set("databaseId", databaseID)
	}
	if recordID != "" {
		query.Set("recordId", recordID)
	}
	reqURL.RawQuery = query.Encode()
	return reqURL, nil
}

func (c *APIClient) performSourceAttempt(reqURL *url.URL, treeID, shortPersonID string, attempt int) (*FactEditData, bool, error) {
	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create request for source page: %w", err)
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Referer", fmt.Sprintf("https://www.ancestry.com/family-tree/person/tree/%s/person/%s/facts", treeID, shortPersonID))
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to fetch source page: %w", err)
		c.log.Printf("[DEBUG] Attempt %d: %v\n", attempt, err)
		return nil, true, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.log.Printf("Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		err = fmt.Errorf("source page request failed with status %d (URL: %s): %s", resp.StatusCode, reqURL.String(), string(body))
		c.log.Printf("[DEBUG] Attempt %d: %v\n", attempt, err)
		if resp.StatusCode >= 500 && resp.StatusCode < 600 {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
			return nil, true, err
		}
		return nil, false, err
	}

	html, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("failed to read source page body: %w", err)
		c.log.Printf("[DEBUG] Attempt %d: %v\n", attempt, err)
		return nil, true, err
	}

	factEditData, err := c.extractFactEditDataFromHTML(string(html))
	if err != nil {
		c.log.Printf("[DEBUG] Attempt %d: %v\n", attempt, err)
		return nil, true, err
	}

	return factEditData, false, nil
}

func (c *APIClient) extractFactEditDataFromHTML(htmlContent string) (*FactEditData, error) {
	if strings.TrimSpace(htmlContent)[0] == '{' {
		var jsonResp map[string]interface{}
		if err := json.Unmarshal([]byte(htmlContent), &jsonResp); err == nil {
			if htmlVal, ok := jsonResp["html"].(string); ok {
				htmlContent = htmlVal
			}
		}
	}

	startMarker := "window.getFactEditData = "
	startIndex := strings.Index(htmlContent, startMarker)
	if startIndex == -1 {
		return nil, fmt.Errorf("could not find window.getFactEditData in HTML content")
	}

	jsonStartIndex := startIndex + len(startMarker)
	jsonEndIndex := strings.Index(htmlContent[jsonStartIndex:], ";</script>")
	if jsonEndIndex == -1 {
		jsonEndIndex = strings.Index(htmlContent[jsonStartIndex:], "</script>")
		if jsonEndIndex == -1 {
			jsonEndIndex = strings.Index(htmlContent[jsonStartIndex:], ";")
			if jsonEndIndex == -1 {
				return nil, fmt.Errorf("could not find end of window.getFactEditData JSON")
			}
		}
	}
	jsonEndIndex += jsonStartIndex

	jsonStr := strings.TrimSpace(htmlContent[jsonStartIndex:jsonEndIndex])

	var factEditData FactEditData
	if err := json.Unmarshal([]byte(jsonStr), &factEditData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal window.getFactEditData JSON: %w", err)
	}
	return &factEditData, nil
}

// GetDiscoveryRecords retrieves records for a given database and person ID
func (c *APIClient) GetDiscoveryRecords(dbID, pId string) (*DiscoveryDiscoveryRecordResponse, error) {
	endpoint := fmt.Sprintf("%s/discoveryui-contentservice/api/records", c.baseURL)
	reqURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	query := reqURL.Query()
	query.Set("dbid", dbID)
	query.Set("r_idx", pId)
	reqURL.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://www.ancestry.com/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.log.Printf("Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var recordResponse DiscoveryDiscoveryRecordResponse
	if err := json.NewDecoder(resp.Body).Decode(&recordResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &recordResponse, nil
}
