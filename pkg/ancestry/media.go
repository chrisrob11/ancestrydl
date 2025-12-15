package ancestry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

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

// DownloadFile downloads a file from a given URL
func (c *APIClient) DownloadFile(fileURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", "http://ancestry.com/"+fileURL, nil)
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

// DownloadRecordImage downloads an image from a RecordImageUrl with security token
// This is the preferred method for downloading census images and other record images
// that require authentication. The recordImageURL should be the URL from PersonSourceDetail.RecordImageUrl
// which includes the security token. This function removes size restrictions to get full-size images.
func (c *APIClient) DownloadRecordImage(recordImageURL string) ([]byte, error) {
	// The recordImageURL is typically a relative URL like:
	// "/api/media/retrieval/v2/image/namespaces/62308/media/43290879-Connecticut-023376-0010.jpg?client=PersonUI&securityToken=xwd2f659e76cf58bfb8201982a2c0435f4e8de3ba50c962c00&maxHeight=250"

	// Parse the URL to modify query parameters
	var fullURL string
	if strings.HasPrefix(recordImageURL, "http") {
		fullURL = recordImageURL
	} else {
		fullURL = c.baseURL + recordImageURL
	}

	reqURL, err := url.Parse(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse record image URL: %w", err)
	}

	// Remove size restrictions to get full-size image
	query := reqURL.Query()
	query.Del("maxWidth")
	query.Del("maxHeight")
	query.Del("maxSide")
	reqURL.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "image/webp,image/apng,image/*,*/*;q=0.8")
	req.Header.Set("Referer", c.baseURL+"/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download record image: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("record image download failed with status %d: %s", resp.StatusCode, string(body))
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read record image data: %w", err)
	}

	return imageData, nil
}
