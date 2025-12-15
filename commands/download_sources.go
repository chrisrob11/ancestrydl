package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/chrisrob11/ancestrydl/pkg/ancestry"
	"github.com/urfave/cli/v2"
)

const unknownPersonName = "Unknown"

// DownloadSources downloads all source records for all people in a tree
func DownloadSources(c *cli.Context) error {
	treeID := c.Args().First()
	if treeID == "" {
		return cli.Exit("Error: tree-id is required", 1)
	}

	outputBaseDir := c.String("output")
	if outputBaseDir == "" {
		outputBaseDir = fmt.Sprintf("./tree-%s-sources", treeID)
	}

	verbose := c.Bool("verbose")

	fmt.Printf("Downloading sources for tree %s to: %s\n", treeID, outputBaseDir)
	if verbose {
		fmt.Println("Verbose mode enabled")
	}
	fmt.Println()

	apiClient, err := setupAPIClientForDownload(verbose)
	if err != nil {
		return err
	}
	defer func() {
		if err := apiClient.Close(); err != nil {
			fmt.Printf("Error closing API client: %v\n", err)
		}
	}()

	allPersons, err := fetchTreePersons(apiClient, treeID)
	if err != nil {
		return err
	}

	// Create output directories
	sourcesDir, peopleSourcesDir, mediaDir, err := createSourceDirectories(outputBaseDir)
	if err != nil {
		return err
	}

	// Map to store unique sources to avoid re-downloading
	downloadedSources := make(map[string]*ancestry.FactEditData)
	peopleWithSources := 0

	fmt.Println("3. Collecting sources for each person...")
	peopleWithSources = processAllPersons(apiClient, treeID, allPersons, downloadedSources, mediaDir, peopleSourcesDir, verbose)

	fmt.Println("5. Saving unique source data files...")
	sourcesSavedCount, totalMediaDownloaded := saveDownloadedSources(downloadedSources, sourcesDir)

	printSourceDownloadSummary(sourcesSavedCount, peopleWithSources, totalMediaDownloaded, sourcesDir, peopleSourcesDir, mediaDir)

	return nil
}

func fetchTreePersons(apiClient *ancestry.APIClient, treeID string) ([]ancestry.Person, error) {
	// 1. Get all people
	fmt.Println("1. Getting person count...")
	totalCount, err := apiClient.GetPersonsCount(treeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get person count: %w", err)
	}
	fmt.Printf("   ✓ Tree has %d persons\n", totalCount)

	fmt.Println("2. Fetching list of people...")
	allPersons, err := downloadAllPersons(apiClient, treeID, totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to download person list: %w", err)
	}
	fmt.Printf("   ✓ Found %d persons\n", len(allPersons))
	return allPersons, nil
}

func createSourceDirectories(outputBaseDir string) (string, string, string, error) {
	sourcesDir := filepath.Join(outputBaseDir, "sources")
	peopleSourcesDir := filepath.Join(outputBaseDir, "people_sources")
	mediaDir := filepath.Join(outputBaseDir, "media") // New media directory
	if err := os.MkdirAll(sourcesDir, 0755); err != nil {
		return "", "", "", fmt.Errorf("failed to create sources directory: %w", err)
	}
	if err := os.MkdirAll(peopleSourcesDir, 0755); err != nil {
		return "", "", "", fmt.Errorf("failed to create people_sources directory: %w", err)
	}
	if err := os.MkdirAll(mediaDir, 0755); err != nil { // Create media directory
		return "", "", "", fmt.Errorf("failed to create media directory: %w", err)
	}
	return sourcesDir, peopleSourcesDir, mediaDir, nil
}

func processAllPersons(apiClient *ancestry.APIClient, treeID string, allPersons []ancestry.Person, downloadedSources map[string]*ancestry.FactEditData, mediaDir, peopleSourcesDir string, verbose bool) int {
	peopleWithSources := 0
	for i, person := range allPersons {
		// Log progress
		if (i+1)%10 == 0 || i == 0 || (i+1) == len(allPersons) {
			personName := person.GetDisplayName()
			if personName == "" {
				personName = unknownPersonName
			}
			fmt.Printf("   Processing person %d/%d: %s...\n", i+1, len(allPersons), personName)
		}

		hasSources, err := processPersonForSources(apiClient, treeID, person, downloadedSources, mediaDir, peopleSourcesDir, verbose)
		if err != nil {
			// Log error but continue
			name := person.GetDisplayName()
			if name == "" {
				name = unknownPersonName
			}
			fmt.Printf("   [Warning] Failed to process %s: %v\n", name, err)
			continue
		}
		if hasSources {
			peopleWithSources++
		}
	}
	return peopleWithSources
}

func printSourceDownloadSummary(sourcesSavedCount, peopleWithSources, totalMediaDownloaded int, sourcesDir, peopleSourcesDir, mediaDir string) {
	fmt.Println()
	fmt.Printf("✅ Downloaded %d unique sources, associated with %d people.\n", sourcesSavedCount, peopleWithSources)
	fmt.Printf("✅ Downloaded %d media files (referenced in sources).\n", totalMediaDownloaded)
	fmt.Printf("   Unique source data saved to: %s\n", sourcesDir)
	fmt.Printf("   Per-person source indexes saved to: %s\n", peopleSourcesDir)
	fmt.Printf("   Media files saved to: %s\n", mediaDir)
}

func processPersonForSources(apiClient *ancestry.APIClient, treeID string, person ancestry.Person, downloadedSources map[string]*ancestry.FactEditData, mediaDir, peopleSourcesDir string, verbose bool) (bool, error) {
	personID := person.GetPersonID()
	personName := person.GetDisplayName()
	if personName == "" {
		personName = unknownPersonName
	}

	// Fetch facts for the person
	researchData, err := apiClient.GetPersonFactsFromHTML(treeID, personID)
	if err != nil {
		return false, fmt.Errorf("failed to get facts for %s (ID: %s): %w", personName, personID, err)
	}

	if researchData == nil || len(researchData.PersonFacts) == 0 {
		if verbose {
			fmt.Printf("      No facts found for %s\n", personName)
		}
		return false, nil
	}

	if verbose {
		fmt.Printf("      Found %d facts for %s\n", len(researchData.PersonFacts), personName)
	}

	citationIDsForPerson := processFacts(researchData, downloadedSources, apiClient, mediaDir, verbose)

	if len(citationIDsForPerson) > 0 {
		if verbose {
			fmt.Printf("      Found %d citations for %s\n", len(citationIDsForPerson), personName)
		}
		savePersonSourceIndex(peopleSourcesDir, personName, personID, citationIDsForPerson)
		return true, nil
	} else if verbose {
		fmt.Printf("      No citations found for %s\n", personName)
	}

	return false, nil
}

func processFacts(researchData *ancestry.ResearchData, downloadedSources map[string]*ancestry.FactEditData, apiClient *ancestry.APIClient, mediaDir string, verbose bool) []string {
	var citationIDsForPerson []string

	uniqueCitationIDsForPerson := make(map[string]bool)

	personSourcesMap := make(map[string]ancestry.PersonSourceDetail)
	if researchData.PersonSources != nil {
		for _, ps := range researchData.PersonSources {
			personSourcesMap[ps.CitationId] = ps
		}
	}

	for _, fact := range researchData.PersonFacts {
		citationIDs := ExtractCitationIDs(fact)
		for _, cid := range citationIDs {
			cid = strings.TrimSpace(cid)
			if cid == "" {
				continue
			}
			if uniqueCitationIDsForPerson[cid] {
				continue
			}
			uniqueCitationIDsForPerson[cid] = true
			citationIDsForPerson = append(citationIDsForPerson, cid)

			if _, ok := downloadedSources[cid]; !ok {
				sourceData := downloadSource(apiClient, personSourcesMap, cid, mediaDir, verbose)
				if sourceData != nil {
					downloadedSources[cid] = sourceData
				}
			}
		}
	}
	return citationIDsForPerson
}

func downloadSource(apiClient *ancestry.APIClient, personSourcesMap map[string]ancestry.PersonSourceDetail, cid, mediaDir string, verbose bool) *ancestry.FactEditData {
	psDetail, found := personSourcesMap[cid]
	if !found {
		if verbose {
			fmt.Printf("      [Warning] No PersonSourceDetail found for citation %s\n", cid)
		}
		return nil
	}

	sourceData := &ancestry.FactEditData{
		CitationID:            psDetail.CitationId,
		DatabaseID:            psDetail.DatabaseId,
		RecordID:              psDetail.RecordId,
		SourceID:              psDetail.SourceId,
		CitationTitle:         psDetail.Title,
		RecordImageUrl:        psDetail.RecordImageUrl,
		RecordImagePreviewUrl: psDetail.RecordImagePreviewUrl,
	}

	if psDetail.RecordImageUrl != "" {
		var writer, errWriter io.Writer
		if verbose {
			writer = os.Stdout
		}
		// Always log errors to stdout if we are in CLI, but reusing existing logic that used printf
		errWriter = os.Stdout

		localPath, _ := DownloadAndSaveRecordImage(writer, errWriter, apiClient, psDetail.RecordImageUrl, cid, mediaDir, "media")
		if localPath != "" {
			sourceData.LocalMediaFilePath = localPath
		}
	}
	return sourceData
}

func saveDownloadedSources(downloadedSources map[string]*ancestry.FactEditData, sourcesDir string) (int, int) {
	sourcesSavedCount := 0
	totalMediaDownloaded := 0
	for cid, sourceData := range downloadedSources {
		sourceFilePath := filepath.Join(sourcesDir, fmt.Sprintf("%s.json", cid))
		jsonData, err := json.MarshalIndent(sourceData, "", "  ")
		if err != nil {
			fmt.Printf("   [Error] Failed to marshal unique source %s: %v\n", cid, err)
			continue
		}
		if err := os.WriteFile(sourceFilePath, jsonData, 0644); err != nil {
			fmt.Printf("   [Error] Failed to save unique source %s: %v\n", cid, err)
			continue
		}
		sourcesSavedCount++
		if sourceData.LocalMediaFilePath != "" {
			totalMediaDownloaded++
		}
	}
	return sourcesSavedCount, totalMediaDownloaded
}

func savePersonSourceIndex(peopleSourcesDir, personName, personID string, citationIDs []string) {
	personSourcesFilePath := filepath.Join(peopleSourcesDir, fmt.Sprintf("%s-%s.json", sanitizeFilename(personName), getShortPersonID(personID)))
	personSourceIndexData, err := json.MarshalIndent(citationIDs, "", "  ")
	if err != nil {
		fmt.Printf("   [Error] Failed to marshal source index for %s: %v\n", personName, err)
		return
	}
	if err := os.WriteFile(personSourcesFilePath, personSourceIndexData, 0644); err != nil {
		fmt.Printf("   [Error] Failed to write source index for %s: %v\n", personName, err)
	}
}
