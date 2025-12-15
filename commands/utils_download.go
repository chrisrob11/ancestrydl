package commands

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/chrisrob11/ancestrydl/pkg/ancestry"
)

const jpgExtension = ".jpg"

var mediaURLRegex = regexp.MustCompile(`/api/media/retrieval/v2/image/namespaces/(\d+)/media/([a-zA-Z0-9_-]+)\.jpg`)

// ExtractMediaDetailsFromURL parses a media URL to extract the namespace and mediaGUID.
func ExtractMediaDetailsFromURL(mediaURL string) (namespace, mediaGUID string, ok bool) {
	matches := mediaURLRegex.FindStringSubmatch(mediaURL)
	if len(matches) == 3 {
		return matches[1], matches[2], true
	}
	return "", "", false
}

// DownloadAndSaveRecordImage downloads a record image and saves it to the media directory.
// It handles filename generation and error logging.
func DownloadAndSaveRecordImage(writer, errWriter io.Writer, client *ancestry.APIClient, recordImageUrl, sourceID, mediaDir, relativePathPrefix string) (string, error) {
	if recordImageUrl == "" {
		return "", nil
	}

	if writer != nil {
		_, _ = fmt.Fprintf(writer, "Downloading record image for source %s...\n", sourceID)
	}

	imageData, err := client.DownloadRecordImage(recordImageUrl)
	if err != nil {
		if errWriter != nil {
			_, _ = fmt.Fprintf(errWriter, "[Warning] Failed to download record image for source %s: %v\n", sourceID, err)
		}
		return "", err
	}

	mediaFileName := fmt.Sprintf("%s_record%s", sourceID, jpgExtension)

	// Extract filename from URL path
	// URL format: /api/media/retrieval/v2/image/namespaces/62308/media/43290879-Connecticut-023376-0010.jpg?params...
	// We want just the filename part: 43290879-Connecticut-023376-0010.jpg
	urlPath := strings.Split(recordImageUrl, "?")[0] // Remove query params
	pathParts := strings.Split(urlPath, "/")
	if len(pathParts) > 0 {
		lastPart := pathParts[len(pathParts)-1]
		if lastPart != "" && strings.Contains(lastPart, ".") {
			// Use the actual filename from URL
			mediaFileName = fmt.Sprintf("%s_%s", sourceID, lastPart)
		}
	}

	mediaFilePath := filepath.Join(mediaDir, mediaFileName)
	if err := os.WriteFile(mediaFilePath, imageData, 0644); err != nil {
		if errWriter != nil {
			_, _ = fmt.Fprintf(errWriter, "[Warning] Failed to save record image for source %s: %v\n", sourceID, err)
		}
		return "", err
	}

	if writer != nil {
		_, _ = fmt.Fprintf(writer, "Successfully downloaded full-size record image for source %s as %s (%d bytes)\n",
			sourceID, mediaFileName, len(imageData))
	}

	return filepath.ToSlash(filepath.Join(relativePathPrefix, mediaFileName)), nil
}

// ExtractCitationIDs parses the SourceCitationIDs field from a PersonFactDetail.
func ExtractCitationIDs(fact ancestry.PersonFactDetail) []string {
	if fact.SourceCitationIDs == nil {
		return nil
	}

	var citationIDs []string
	switch v := fact.SourceCitationIDs.(type) {
	case string:
		if v != "" {
			cleaned := strings.ReplaceAll(v, ",", " ")
			citationIDs = strings.Fields(cleaned)
		}
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok {
				cleaned := strings.ReplaceAll(s, ",", " ")
				fields := strings.Fields(cleaned)
				citationIDs = append(citationIDs, fields...)
			}
		}
	}
	return citationIDs
}

// DetectFileExtension detects file extension from file data using magic bytes
func DetectFileExtension(data []byte) string {
	if len(data) < 4 { // Minimal length for most magic bytes
		return ".bin"
	}

	// JPEG
	if bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}) {
		return jpgExtension
	}

	// PNG
	if bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47}) {
		return ".png"
	}

	// GIF
	if bytes.HasPrefix(data, []byte("GIF89a")) || bytes.HasPrefix(data, []byte("GIF87a")) {
		return ".gif"
	}

	// WebP
	if len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
		return ".webp"
	}

	// PDF
	if bytes.HasPrefix(data, []byte("%PDF-")) {
		return ".pdf"
	}

	// Default to jpg for images (most common from Ancestry)
	return jpgExtension
}
