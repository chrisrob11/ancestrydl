package ancestry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

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
