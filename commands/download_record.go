package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chrisrob11/ancestrydl/pkg/ancestry"
	"github.com/chrisrob11/ancestrydl/pkg/config"
	"github.com/urfave/cli/v2"
)

// DownloadRecord downloads a specific record for a person in a tree.
func DownloadRecord(c *cli.Context) error {
	recordTreeID := c.String("tree-id")
	recordpID := c.String("person-id")
	sourceID := c.String("source-id")

	if recordTreeID == "" || recordpID == "" {
		return cli.Exit("Error: tree-id and person-id are required", 1)
	}

	client, err := setupClient(c)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	// Always fetch person facts to get source details (databaseId, recordId)
	_, _ = fmt.Fprintf(c.App.Writer, "Fetching facts for person %s in tree %s...\n", recordpID, recordTreeID)
	researchData, err := client.GetPersonFactsFromHTML(recordTreeID, recordpID)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Error getting person facts: %v", err), 1)
	}

	if researchData == nil {
		return cli.Exit("No research data found for person", 1)
	}

	personSourcesMap := make(map[string]ancestry.PersonSourceDetail)
	if researchData.PersonSources != nil {
		for _, ps := range researchData.PersonSources {
			personSourcesMap[ps.CitationId] = ps
		}
	}

	mediaDir, err := createMediaDir(recordTreeID)
	if err != nil {
		return cli.Exit(err.Error(), 1)
	}

	if sourceID != "" {
		return downloadSingleSource(c, client, personSourcesMap, sourceID, recordpID, mediaDir)
	}

	return downloadAllSources(c, researchData, personSourcesMap, client, mediaDir)
}

func setupClient(c *cli.Context) (*ancestry.APIClient, error) {
	cookiesJSON, err := config.GetCookies()
	if err != nil {
		return nil, cli.Exit(fmt.Sprintf("Error reading cookies: %v", err), 1)
	}

	cookies, err := ancestry.DeserializeCookies(cookiesJSON)
	if err != nil {
		return nil, cli.Exit(fmt.Sprintf("Error deserializing cookies: %v", err), 1)
	}

	client, err := ancestry.NewAPIClient(cookies, c.Bool("verbose"))
	if err != nil {
		return nil, cli.Exit(fmt.Sprintf("Error creating API client: %v", err), 1)
	}
	return client, nil
}

func createMediaDir(recordTreeID string) (string, error) {
	outputBaseDir := fmt.Sprintf("./tree-%s-sources", recordTreeID)
	mediaDir := filepath.Join(outputBaseDir, "media")
	if err := os.MkdirAll(mediaDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create media directory: %v", err)
	}
	return mediaDir, nil
}

func downloadSingleSource(c *cli.Context, client *ancestry.APIClient, personSourcesMap map[string]ancestry.PersonSourceDetail, sourceID, recordpID, mediaDir string) error {
	psDetail, found := personSourcesMap[sourceID]
	if !found {
		return cli.Exit(fmt.Sprintf("Error: source %s not found in PersonSources for person %s", sourceID, recordpID), 1)
	}

	sourceData := createSourceData(psDetail)

	localPath, _ := DownloadAndSaveRecordImage(c.App.Writer, c.App.ErrWriter, client, psDetail.RecordImageUrl, sourceID, mediaDir, "media")
	if localPath != "" {
		sourceData.LocalMediaFilePath = localPath
	}

	jsonBytes, err := json.MarshalIndent(sourceData, "", "  ")
	if err != nil {
		return cli.Exit(fmt.Sprintf("Error marshalling source data to JSON: %v", err), 1)
	}
	fmt.Println(string(jsonBytes))

	return nil
}

func downloadAllSources(c *cli.Context, researchData *ancestry.ResearchData, personSourcesMap map[string]ancestry.PersonSourceDetail, client *ancestry.APIClient, mediaDir string) error {
	var allSources []*ancestry.FactEditData
	uniqueCitationIDs := make(map[string]bool)
	totalMediaDownloaded := 0

	for _, fact := range researchData.PersonFacts {
		citationIDs := ExtractCitationIDs(fact)
		for _, cid := range citationIDs {
			cid = strings.TrimSpace(cid)
			if cid == "" || uniqueCitationIDs[cid] {
				continue
			}
			uniqueCitationIDs[cid] = true

			_, _ = fmt.Fprintf(c.App.Writer, "Downloading source %s...\n", cid)

			psDetail, found := personSourcesMap[cid]
			if !found {
				_, _ = fmt.Fprintf(c.App.ErrWriter, "[Warning] No PersonSourceDetail found for citation %s\n", cid)
				continue
			}

			sourceData := createSourceData(psDetail)

			if psDetail.RecordImageUrl != "" {
				localPath, err := DownloadAndSaveRecordImage(c.App.Writer, c.App.ErrWriter, client, psDetail.RecordImageUrl, cid, mediaDir, "media")
				if err == nil && localPath != "" {
					sourceData.LocalMediaFilePath = localPath
					totalMediaDownloaded++
				}
			}

			allSources = append(allSources, sourceData)
		}
	}

	jsonBytes, err := json.MarshalIndent(allSources, "", "  ")
	if err != nil {
		return cli.Exit(fmt.Sprintf("Error marshalling sources to JSON: %v", err), 1)
	}
	fmt.Println(string(jsonBytes))
	_, _ = fmt.Fprintf(c.App.Writer, "✅ Downloaded %d media files to %s.\n", totalMediaDownloaded, mediaDir)
	return nil
}

func createSourceData(psDetail ancestry.PersonSourceDetail) *ancestry.FactEditData {
	return &ancestry.FactEditData{
		CitationID:            psDetail.CitationId,
		DatabaseID:            psDetail.DatabaseId,
		RecordID:              psDetail.RecordId,
		SourceID:              psDetail.SourceId,
		CitationTitle:         psDetail.Title,
		RecordImageUrl:        psDetail.RecordImageUrl,
		RecordImagePreviewUrl: psDetail.RecordImagePreviewUrl,
	}
}
