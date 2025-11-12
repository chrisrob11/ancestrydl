package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chrisrob11/ancestrydl/pkg/ancestry"
	"github.com/chrisrob11/ancestrydl/pkg/config"
	"github.com/urfave/cli/v2"
)

const (
	// Birth is a constant for the string "Birth"
	Birth = "Birth"
	// Death is a constant for the string "Death"
	Death = "Death"
)

// TreeExport represents the complete tree export structure
type TreeExport struct {
	TreeID      string             `json:"treeId"`
	TreeName    string             `json:"treeName"`
	ExportDate  string             `json:"exportDate"`
	PersonCount int                `json:"personCount"`
	Persons     []ancestry.Person  `json:"persons"`
	TreeInfo    *ancestry.TreeInfo `json:"treeInfo,omitempty"`
}

// extractPlaceFromNPS extracts place name from Nested Place Structure
// NPS is an array of maps containing place information
func extractPlaceFromNPS(nps []map[string]interface{}) string {
	if len(nps) == 0 {
		return ""
	}

	// NPS typically contains place components like city, county, state, country
	// We want to build a comma-separated string from the place names
	var places []string
	for _, placeMap := range nps {
		// Check for 'v' field which contains the place name value
		if v, ok := placeMap["v"].(string); ok && v != "" {
			places = append(places, v)
		}
	}

	return strings.Join(places, ", ")
}

// getTreeIDForDownload retrieves tree ID from arguments or uses default
func getTreeIDForDownload(c *cli.Context) (string, error) {
	treeID := c.Args().First()
	if treeID != "" {
		return treeID, nil
	}

	defaultTreeID, err := config.GetDefaultTreeID()
	if err != nil {
		return "", fmt.Errorf("failed to get default tree: %w", err)
	}
	if defaultTreeID == "" {
		return "", fmt.Errorf("tree ID is required\n\nUsage: ancestrydl download-tree <tree-id> --output <directory>")
	}
	fmt.Printf("Using default tree: %s\n", defaultTreeID)
	return defaultTreeID, nil
}

// setupAPIClientForDownload creates an API client from stored cookies
func setupAPIClientForDownload(verbose bool) (*ancestry.APIClient, error) {
	cookiesJSON, err := config.GetCookies()
	if err != nil {
		return nil, fmt.Errorf("failed to load stored cookies: %w\n\nPlease run 'ancestrydl login' first to authenticate", err)
	}

	fmt.Println("1. Creating API client...")
	apiClient, err := ancestry.NewAPIClientFromJSON(cookiesJSON, verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}
	fmt.Println("   ✓ API client ready")
	return apiClient, nil
}

// fetchTreeData downloads all persons, relationships, and events from the tree
func fetchTreeData(apiClient *ancestry.APIClient, treeID string) ([]ancestry.Person, map[string]PersonRelationship, int, error) {
	fmt.Println("3. Getting person count...")
	totalCount, err := apiClient.GetPersonsCount(treeID)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get person count: %w", err)
	}
	fmt.Printf("   ✓ Tree has %d persons\n", totalCount)

	fmt.Println("4. Downloading all persons...")
	allPersons, err := downloadAllPersons(apiClient, treeID, totalCount)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to download persons: %w", err)
	}
	fmt.Printf("   ✓ Downloaded %d persons\n", len(allPersons))

	fmt.Println("5. Building relationship map...")
	relationships, familyViewEvents := buildRelationships(apiClient, treeID, allPersons)
	fmt.Printf("   ✓ Built relationships for %d persons\n", len(relationships))

	// Merge FamilyView events into persons
	for i := range allPersons {
		personID := allPersons[i].GetPersonID()
		if events, hasEvents := familyViewEvents[personID]; hasEvents && len(events) > 0 {
			allPersons[i].Events = events
		}
	}

	fmt.Println("6. Fetching complete event data from Facts pages...")
	fetchFactsForAllPersons(apiClient, treeID, allPersons)
	fmt.Println("   ✓ Fetched complete event data")

	fmt.Println("7. Inferring event types from relationships...")
	inferredCount := inferEventTypes(allPersons, relationships)
	fmt.Printf("   ✓ Inferred %d event types\n", inferredCount)

	return allPersons, relationships, totalCount, nil
}

// saveTreeOutput saves all tree data, media, and generates the HTML viewer
func saveTreeOutput(apiClient *ancestry.APIClient, treeID, outputDir string, treeInfo *ancestry.TreeInfo,
	allPersons []ancestry.Person, relationships map[string]PersonRelationship) (int, error) {
	fmt.Println("8. Creating output directories...")
	if err := createDirectoryStructure(outputDir); err != nil {
		return 0, fmt.Errorf("failed to create directories: %w", err)
	}
	fmt.Println("   ✓ Directories created")

	fmt.Println("9. Downloading media files...")
	mediaIndex, downloadCount := downloadAllMedia(apiClient, treeID, allPersons, outputDir)
	fmt.Printf("   ✓ Downloaded %d media files\n", downloadCount)

	fmt.Println("10. Saving tree data...")
	treeExport := TreeExport{
		TreeID:      treeID,
		TreeName:    treeInfo.TreeName,
		ExportDate:  time.Now().Format(time.RFC3339),
		PersonCount: len(allPersons),
		Persons:     allPersons,
		TreeInfo:    treeInfo,
	}

	if err := saveTreeData(outputDir, &treeExport, relationships, mediaIndex); err != nil {
		return 0, fmt.Errorf("failed to save tree data: %w", err)
	}
	fmt.Println("   ✓ Tree data saved")

	fmt.Println("11. Generating HTML viewer...")
	if err := generateHTMLViewer(outputDir, &treeExport); err != nil {
		fmt.Printf("   Warning: Failed to generate HTML viewer: %v\n", err)
	} else {
		fmt.Println("   ✓ HTML viewer created")
	}

	return downloadCount, nil
}

// printDownloadSummary prints the summary of downloaded tree data
func printDownloadSummary(outputDir string, downloadCount int) {
	fmt.Println("\n✅ Tree download complete!")
	fmt.Printf("   Output: %s\n", outputDir)
	fmt.Println()
	fmt.Println("Files created:")
	fmt.Println("  • index.html - Interactive HTML viewer (open directly in browser)")
	fmt.Println("  • people.json - All persons with readable details")
	fmt.Println("  • metadata.json - Tree information")
	if downloadCount > 0 {
		fmt.Printf("  • media/ - %d media files (photos, documents)\n", downloadCount)
		fmt.Println("  • media-index.json - Media file index with titles and descriptions")
	}
	fmt.Println()
	fmt.Printf("👉 To view your tree, open: %s/index.html\n", outputDir)
	fmt.Println()
}

// DownloadTree downloads a complete family tree with all data and media
func DownloadTree(c *cli.Context) error {
	treeID, err := getTreeIDForDownload(c)
	if err != nil {
		return err
	}

	outputDir := c.String("output")
	if outputDir == "" {
		outputDir = fmt.Sprintf("./tree-%s", treeID)
	}

	verbose := c.Bool("verbose")

	fmt.Printf("Downloading tree %s to: %s\n", treeID, outputDir)
	if verbose {
		fmt.Println("Verbose mode enabled: HTTP requests/responses will be logged to http_log.txt")
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

	fmt.Println("2. Fetching tree information...")
	treeInfo, err := apiClient.GetTreeInfo(treeID)
	if err != nil {
		fmt.Printf("   Warning: Could not fetch tree info: %v\n", err)
	} else {
		fmt.Printf("   ✓ Tree: %s\n", treeInfo.TreeName)
	}

	allPersons, relationships, _, err := fetchTreeData(apiClient, treeID)
	if err != nil {
		return err
	}

	downloadCount, err := saveTreeOutput(apiClient, treeID, outputDir, treeInfo, allPersons, relationships)
	if err != nil {
		return err
	}

	printDownloadSummary(outputDir, downloadCount)

	return nil
}

// PersonRelationship stores relationship information for a person
type PersonRelationship struct {
	PersonID string                  `json:"personId"`
	Name     string                  `json:"name"`
	Parents  []RelationshipReference `json:"parents,omitempty"`
	Spouses  []RelationshipReference `json:"spouses,omitempty"`
	Children []RelationshipReference `json:"children,omitempty"`
}

// RelationshipReference references another person in the tree
type RelationshipReference struct {
	PersonID string `json:"personId"`
	Name     string `json:"name"`
}

// extractPersonNumber extracts the person number from a full person ID
func extractPersonNumber(personID string) string {
	if idx := strings.Index(personID, ":"); idx > 0 {
		return personID[:idx]
	}
	return personID
}

// buildFamilyViewPersonsMap creates a map from person ID to Person pointer
func buildFamilyViewPersonsMap(persons []ancestry.Person) map[string]*ancestry.Person {
	personsMap := make(map[string]*ancestry.Person)
	for j := range persons {
		p := &persons[j]
		pid := p.GetPersonID()
		if pid != "" {
			personsMap[pid] = p
		}
	}
	return personsMap
}

// extractTargetIDAndName extracts the target person's ID and name from a family member
func extractTargetIDAndName(familyMember ancestry.FamilyMember, personsMap map[string]*ancestry.Person) (string, string, bool) {
	targetID, ok := familyMember.TGID["v"].(string)
	if !ok {
		return "", "", false
	}

	targetName := targetID
	if targetPerson, ok := personsMap[targetID]; ok {
		targetName = targetPerson.GetDisplayName()
	}

	return targetID, targetName, true
}

// addFamilyMemberToRelationship adds a family member to the appropriate relationship list
func addFamilyMemberToRelationship(rel *PersonRelationship, familyMember ancestry.FamilyMember, ref RelationshipReference) {
	switch familyMember.Type {
	case "F", "M": // Father or Mother
		rel.Parents = append(rel.Parents, ref)
	case "H", "W": // Husband or Wife (Spouse)
		rel.Spouses = append(rel.Spouses, ref)
	case "C": // Child
		rel.Children = append(rel.Children, ref)
	}
}

// processFamilyView processes a person's family view to extract relationships
func processFamilyView(personID string, familyView *ancestry.FamilyViewResponse) (PersonRelationship, []ancestry.Event, bool) {
	personsMap := buildFamilyViewPersonsMap(familyView.Persons)

	focusPerson, exists := personsMap[personID]
	if !exists {
		return PersonRelationship{}, nil, false
	}

	rel := PersonRelationship{
		PersonID: personID,
		Name:     focusPerson.GetDisplayName(),
		Parents:  []RelationshipReference{},
		Spouses:  []RelationshipReference{},
		Children: []RelationshipReference{},
	}

	for _, familyMember := range focusPerson.Family {
		targetID, targetName, ok := extractTargetIDAndName(familyMember, personsMap)
		if !ok {
			continue
		}

		ref := RelationshipReference{
			PersonID: targetID,
			Name:     targetName,
		}

		addFamilyMemberToRelationship(&rel, familyMember, ref)
	}

	return rel, focusPerson.Events, true
}

// buildRelationships creates a map of relationships for all persons
// It also returns a map of person IDs to their Events from FamilyView API (which has more complete data)
func buildRelationships(apiClient *ancestry.APIClient, treeID string, persons []ancestry.Person) (map[string]PersonRelationship, map[string][]ancestry.Event) {
	relationships := make(map[string]PersonRelationship)
	eventsMap := make(map[string][]ancestry.Event)

	for i, person := range persons {
		personID := person.GetPersonID()
		if personID == "" {
			continue
		}

		if i%10 == 0 && i > 0 {
			fmt.Printf("   Building relationships %d/%d...\n", i, len(persons))
		}

		personNumber := extractPersonNumber(personID)

		familyView, err := apiClient.GetFamilyView(treeID, personNumber, 1, 1)
		if err != nil {
			if i < 3 {
				fmt.Printf("   [Debug] Failed to get family view for %s: %v\n", person.GetDisplayName(), err)
			}
			continue
		}

		rel, events, ok := processFamilyView(personID, familyView)
		if !ok {
			continue
		}

		relationships[personID] = rel
		if len(events) > 0 {
			eventsMap[personID] = events
		}
	}

	return relationships, eventsMap
}

// downloadAllPersons fetches all persons from the tree with pagination
func downloadAllPersons(apiClient *ancestry.APIClient, treeID string, totalCount int) ([]ancestry.Person, error) {
	limit := 100
	totalPages := (totalCount + limit - 1) / limit

	allPersons := []ancestry.Person{}

	for page := 1; page <= totalPages; page++ {
		fmt.Printf("   Fetching page %d/%d...\n", page, totalPages)
		persons, err := apiClient.GetAllPersons(treeID, page, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch page %d: %w", page, err)
		}
		allPersons = append(allPersons, persons...)
	}

	return allPersons, nil
}

// createDirectoryStructure creates the output directory structure
func createDirectoryStructure(outputDir string) error {
	dirs := []string{
		outputDir,
		filepath.Join(outputDir, "media"),
		filepath.Join(outputDir, "media", "photos"),
		filepath.Join(outputDir, "media", "documents"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// saveTreeData saves all tree data to JSON files
// convertEventToReadableFormat converts an ancestry event to readable map format
func convertEventToReadableFormat(event ancestry.Event) map[string]interface{} {
	eventData := map[string]interface{}{
		"type": event.Type,
		"date": event.Date,
	}

	if place := extractPlaceFromNPS(event.NPS); place != "" {
		eventData["place"] = place
	}

	if event.Description != "" {
		eventData["description"] = event.Description
	}

	return eventData
}

// convertPersonToReadableFormat converts a person to a readable map with relationships and media
func convertPersonToReadableFormat(person ancestry.Person, relationships map[string]PersonRelationship,
	mediaIndex map[string]PersonMediaInfo) map[string]interface{} {
	personID := person.GetPersonID()
	readable := map[string]interface{}{
		"personId": personID,
		"fullName": person.GetDisplayName(),
		"isLiving": person.IsLiving,
	}

	// Add name details
	if len(person.Names) > 0 {
		readable["givenName"] = person.Names[0].GivenName
		readable["surname"] = person.Names[0].Surname
	}

	// Add gender if present
	if person.Gender != "" {
		readable["gender"] = person.Gender
	}

	// Add events in readable format
	if len(person.Events) > 0 {
		events := make([]map[string]interface{}, 0, len(person.Events))
		for _, event := range person.Events {
			events = append(events, convertEventToReadableFormat(event))
		}
		readable["events"] = events
	}

	// Add relationships
	if rel, hasRels := relationships[personID]; hasRels {
		if len(rel.Parents) > 0 {
			readable["parents"] = rel.Parents
		}
		if len(rel.Spouses) > 0 {
			readable["spouses"] = rel.Spouses
		}
		if len(rel.Children) > 0 {
			readable["children"] = rel.Children
		}
	}

	// Add media files
	if mediaInfo, hasMedia := mediaIndex[personID]; hasMedia && len(mediaInfo.Files) > 0 {
		readable["media"] = mediaInfo.Files
	}

	return readable
}

// savePersonsData saves persons to a JSON file in readable format
func savePersonsData(outputDir string, persons []ancestry.Person, relationships map[string]PersonRelationship,
	mediaIndex map[string]PersonMediaInfo) error {
	readablePersons := make([]map[string]interface{}, 0, len(persons))
	for _, person := range persons {
		readablePersons = append(readablePersons, convertPersonToReadableFormat(person, relationships, mediaIndex))
	}

	readableJSON, err := json.MarshalIndent(readablePersons, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal readable persons data: %w", err)
	}

	personsPath := filepath.Join(outputDir, "people.json")
	if err := os.WriteFile(personsPath, readableJSON, 0644); err != nil {
		return fmt.Errorf("failed to write people.json: %w", err)
	}

	return nil
}

// saveMetadata saves tree metadata to a JSON file
func saveMetadata(outputDir string, treeExport *TreeExport) error {
	metadata := map[string]interface{}{
		"treeId":      treeExport.TreeID,
		"treeName":    treeExport.TreeName,
		"exportDate":  treeExport.ExportDate,
		"personCount": treeExport.PersonCount,
		"treeInfo":    treeExport.TreeInfo,
	}

	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	metadataPath := filepath.Join(outputDir, "metadata.json")
	if err := os.WriteFile(metadataPath, metadataJSON, 0644); err != nil {
		return fmt.Errorf("failed to write metadata.json: %w", err)
	}

	return nil
}

func saveTreeData(outputDir string, treeExport *TreeExport, relationships map[string]PersonRelationship, mediaIndex map[string]PersonMediaInfo) error {
	if err := savePersonsData(outputDir, treeExport.Persons, relationships, mediaIndex); err != nil {
		return err
	}

	if err := saveMetadata(outputDir, treeExport); err != nil {
		return err
	}

	return nil
}

// PersonMediaInfo tracks media files for a person
type PersonMediaInfo struct {
	PersonID   string          `json:"personId"`
	PersonName string          `json:"personName"`
	Files      []MediaFileInfo `json:"files"`
}

// MediaFileInfo contains information about a downloaded media file
type MediaFileInfo struct {
	FilePath    string `json:"filePath"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Subcategory string `json:"subcategory"`
	Description string `json:"description"`
	Date        string `json:"date"`
	Type        string `json:"type"`
}

// fetchFactsForAllPersons fetches complete event data from Facts pages for all persons
// This includes place names and descriptions that aren't available in the JSON APIs
func fetchFactsForAllPersons(apiClient *ancestry.APIClient, treeID string, persons []ancestry.Person) {
	totalPersons := len(persons)

	for i := range persons {
		personID := persons[i].GetPersonID()

		// Show progress every 10 people
		if (i+1)%10 == 0 || i == 0 {
			fmt.Printf("   Fetching facts %d/%d...\n", i+1, totalPersons)
		}

		// Fetch facts from HTML page
		researchData, err := apiClient.GetPersonFactsFromHTML(treeID, personID)
		if err != nil {
			// Don't fail the whole process, just log and continue
			fmt.Printf("\n   [Warning] Failed to get facts for %s: %v\n", persons[i].GetDisplayName(), err)
			continue
		}

		if researchData == nil || len(researchData.PersonFacts) == 0 {
			continue
		}

		// Convert PersonFacts to Events and update the person
		events := make([]ancestry.Event, 0, len(researchData.PersonFacts))
		for _, fact := range researchData.PersonFacts {
			// Only include facts that have meaningful data
			if fact.TypeString == "" && fact.Place == "" && fact.Description == "" {
				continue
			}

			// Use Title field for custom events (like "Prison"), otherwise use TypeString
			eventType := fact.TypeString
			if fact.TypeString == "CustomEvent" && fact.Title != "" {
				eventType = fact.Title
			}

			event := ancestry.Event{
				Type:        eventType,
				Date:        fact.Date,
				Description: fact.Description,
			}

			// Add place data if available
			if fact.Place != "" {
				// Create NPS structure to match existing Event format
				nps := []map[string]interface{}{
					{"v": fact.Place},
				}
				event.NPS = nps
			}

			events = append(events, event)
		}

		// Update person's events with complete data
		if len(events) > 0 {
			persons[i].Events = events
		}
	}
}

// getRelationshipGenderLabel returns a gender-specific relationship label
func getRelationshipGenderLabel(gender, maleLabel, femaleLabel, neutralLabel string) string {
	if gender == "m" {
		return maleLabel
	} else if gender == "f" {
		return femaleLabel
	}
	return neutralLabel
}

// processChildEvents processes children's birth and death events for inference
func processChildEvents(childRefs []RelationshipReference, personMap map[string]*ancestry.Person, dateToEventType map[string]string) {
	for _, childRef := range childRefs {
		child, ok := personMap[childRef.PersonID]
		if !ok {
			continue
		}
		for _, evt := range child.Events {
			if (evt.Type == Birth || evt.Type == Death) && evt.Date != nil {
				dateStr := fmt.Sprintf("%v", evt.Date)
				genderLabel := getRelationshipGenderLabel(child.Gender, "son", "daughter", "child")
				label := fmt.Sprintf("%s of %s %s", evt.Type, genderLabel, child.GetDisplayName())
				dateToEventType[dateStr] = label
			}
		}
	}
}

// hasSharedParent checks if two people share at least one parent
func hasSharedParent(myParents, theirParents []RelationshipReference) bool {
	for _, myParent := range myParents {
		for _, theirParent := range theirParents {
			if myParent.PersonID == theirParent.PersonID {
				return true
			}
		}
	}
	return false
}

// processSiblingEvents processes siblings' birth and death events for inference
func processSiblingEvents(personID string, rels PersonRelationship, persons []ancestry.Person,
	relationships map[string]PersonRelationship, dateToEventType map[string]string) {
	if len(rels.Parents) == 0 {
		return
	}

	for _, otherPerson := range persons {
		if otherPerson.GetPersonID() == personID {
			continue
		}
		otherRels, hasOtherRels := relationships[otherPerson.GetPersonID()]
		if !hasOtherRels || len(otherRels.Parents) == 0 {
			continue
		}

		if hasSharedParent(rels.Parents, otherRels.Parents) {
			for _, evt := range otherPerson.Events {
				if (evt.Type == Birth || evt.Type == Death) && evt.Date != nil {
					dateStr := fmt.Sprintf("%v", evt.Date)
					genderLabel := getRelationshipGenderLabel(otherPerson.Gender, "brother", "sister", "sibling")
					label := fmt.Sprintf("%s of %s %s", evt.Type, genderLabel, otherPerson.GetDisplayName())
					dateToEventType[dateStr] = label
				}
			}
		}
	}
}

// processRelativeDeathEvents processes death events for parents or spouses
func processRelativeDeathEvents(relativeRefs []RelationshipReference, personMap map[string]*ancestry.Person,
	dateToEventType map[string]string, maleLabel, femaleLabel, neutralLabel string) {
	for _, relativeRef := range relativeRefs {
		relative, ok := personMap[relativeRef.PersonID]
		if !ok {
			continue
		}
		for _, evt := range relative.Events {
			if evt.Type == Death && evt.Date != nil {
				dateStr := fmt.Sprintf("%v", evt.Date)
				genderLabel := getRelationshipGenderLabel(relative.Gender, maleLabel, femaleLabel, neutralLabel)
				label := fmt.Sprintf("Death of %s %s", genderLabel, relative.GetDisplayName())
				dateToEventType[dateStr] = label
			}
		}
	}
}

// buildPersonMap creates a map from person ID to person pointer for quick lookup
func buildPersonMap(persons []ancestry.Person) map[string]*ancestry.Person {
	personMap := make(map[string]*ancestry.Person)
	for i := range persons {
		personMap[persons[i].GetPersonID()] = &persons[i]
	}
	return personMap
}

// updateEmptyEvents updates empty events with inferred types from the date map
func updateEmptyEvents(person *ancestry.Person, dateToEventType map[string]string) int {
	count := 0
	for j := range person.Events {
		if person.Events[j].Type == "" && person.Events[j].Date != nil {
			dateStr := fmt.Sprintf("%v", person.Events[j].Date)
			if inferredType, found := dateToEventType[dateStr]; found {
				person.Events[j].Type = inferredType
				count++
			}
		}
	}
	return count
}

// inferEventTypes infers event types for empty events based on relationships
// Returns the count of events that were inferred
func inferEventTypes(persons []ancestry.Person, relationships map[string]PersonRelationship) int {
	personMap := buildPersonMap(persons)
	inferredCount := 0

	for i := range persons {
		personID := persons[i].GetPersonID()
		rels, hasRels := relationships[personID]
		if !hasRels {
			continue
		}

		dateToEventType := make(map[string]string)

		// Process different types of relatives
		processChildEvents(rels.Children, personMap, dateToEventType)
		processSiblingEvents(personID, rels, persons, relationships, dateToEventType)
		processRelativeDeathEvents(rels.Parents, personMap, dateToEventType, "father", "mother", "parent")
		processRelativeDeathEvents(rels.Spouses, personMap, dateToEventType, "husband", "wife", "spouse")

		inferredCount += updateEmptyEvents(&persons[i], dateToEventType)
	}

	return inferredCount
}

// getFileExtension extracts extension from URL, removing query parameters
func getFileExtension(url string) string {
	urlPath := url
	if qPos := strings.Index(urlPath, "?"); qPos != -1 {
		urlPath = urlPath[:qPos]
	}
	ext := filepath.Ext(urlPath)
	if ext == "" {
		ext = ".jpg"
	}
	return ext
}

// getShortPersonID returns person ID without colon-separated segments
func getShortPersonID(personID string) string {
	if colonPos := strings.Index(personID, ":"); colonPos != -1 {
		return personID[:colonPos]
	}
	return personID
}

// generateMediaFilename creates a readable filename for media items
func generateMediaFilename(personName, personID string, mediaItem ancestry.PrimaryMediaItem, idx int, ext string) string {
	shortPersonID := getShortPersonID(personID)
	safeName := sanitizeFilename(personName)
	if safeName == "" {
		safeName = "unknown"
	}

	if mediaItem.Subcategory != "" {
		safeSubcategory := sanitizeFilename(mediaItem.Subcategory)
		return fmt.Sprintf("%s-%s-%s-%03d%s", safeName, shortPersonID, safeSubcategory, idx+1, ext)
	}
	if mediaItem.Title != "" {
		safeTitle := sanitizeFilename(mediaItem.Title)
		return fmt.Sprintf("%s-%s-%s-%03d%s", safeName, shortPersonID, safeTitle, idx+1, ext)
	}
	return fmt.Sprintf("%s-%s-%03d%s", safeName, shortPersonID, idx+1, ext)
}

// getMediaSubdirectory determines subdirectory based on media category
func getMediaSubdirectory(category string) string {
	if category == "document" || category == "story" {
		return "documents"
	}
	return "photos"
}

// processMediaItem downloads and saves a single media item
func processMediaItem(apiClient *ancestry.APIClient, mediaItem ancestry.PrimaryMediaItem, personID, personName string,
	idx int, outputDir string) (MediaFileInfo, bool, error) {
	ext := getFileExtension(mediaItem.URL)
	filename := generateMediaFilename(personName, personID, mediaItem, idx, ext)
	subdir := getMediaSubdirectory(mediaItem.Category)

	filePath := filepath.Join(outputDir, "media", subdir, filename)
	relativeFilePath := filepath.Join("media", subdir, filename)

	mediaFileInfo := MediaFileInfo{
		FilePath:    relativeFilePath,
		Title:       mediaItem.Title,
		Category:    mediaItem.Category,
		Subcategory: mediaItem.Subcategory,
		Description: mediaItem.Description,
		Date:        mediaItem.Date,
		Type:        mediaItem.Type,
	}

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return mediaFileInfo, false, nil
	}

	// Download the file
	fileData, err := apiClient.DownloadFile(mediaItem.URL)
	if err != nil {
		return mediaFileInfo, false, fmt.Errorf("download failed: %w", err)
	}

	// Save the file
	if err := os.WriteFile(filePath, fileData, 0644); err != nil {
		return mediaFileInfo, false, fmt.Errorf("save failed: %w", err)
	}

	return mediaFileInfo, true, nil
}

// processPersonMedia fetches and downloads all media for a single person
func processPersonMedia(apiClient *ancestry.APIClient, treeID string, person ancestry.Person,
	outputDir string) (PersonMediaInfo, int, error) {
	personID := person.GetPersonID()
	personName := person.GetDisplayName()
	if personName == "" {
		personName = personID
	}

	personInfo := PersonMediaInfo{
		PersonID:   personID,
		PersonName: personName,
		Files:      []MediaFileInfo{},
	}
	downloaded := 0

	mediaItems, err := apiClient.GetPersonMediaFromAPI(treeID, personID)
	if err != nil {
		return personInfo, 0, fmt.Errorf("error getting media: %w", err)
	}

	if len(mediaItems) == 0 {
		return personInfo, 0, nil
	}

	fmt.Printf("   ✓ Found %d media item(s) for %s (ID: %s)\n",
		len(mediaItems), personName, personID)

	for idx, mediaItem := range mediaItems {
		mediaFileInfo, wasDownloaded, err := processMediaItem(apiClient, mediaItem, personID, personName, idx, outputDir)
		if err != nil {
			fmt.Printf("   [Warning] Failed to process media for %s (ID: %s): %v\n",
				personName, personID, err)
			continue
		}
		personInfo.Files = append(personInfo.Files, mediaFileInfo)
		if wasDownloaded {
			downloaded++
		}
	}

	return personInfo, downloaded, nil
}

// downloadAllMedia downloads all media files for all persons
func downloadAllMedia(apiClient *ancestry.APIClient, treeID string, persons []ancestry.Person, outputDir string) (map[string]PersonMediaInfo, int) {
	mediaIndex := make(map[string]PersonMediaInfo)
	totalDownloaded := 0
	skippedCount := 0

	for i, person := range persons {
		personID := person.GetPersonID()
		personName := person.GetDisplayName()
		if personName == "" {
			personName = personID
		}

		if personID == "" {
			fmt.Printf("   [Debug] Skipping person %d with missing ID (Name: %s, GID: %+v)\n",
				i+1, personName, person.GID)
			skippedCount++
			continue
		}

		if i%10 == 0 {
			fmt.Printf("   Processing person %d/%d (ID: %s, Name: %s)...\n",
				i+1, len(persons), personID, personName)
		}

		personInfo, downloaded, err := processPersonMedia(apiClient, treeID, person, outputDir)
		if err != nil {
			fmt.Printf("   [Warning] %v\n", err)
			continue
		}

		if len(personInfo.Files) > 0 {
			mediaIndex[personID] = personInfo
		}
		totalDownloaded += downloaded
	}

	if skippedCount > 0 {
		fmt.Printf("   Skipped %d persons due to missing person ID\n", skippedCount)
	}

	return mediaIndex, totalDownloaded
}

// sanitizeFilename removes or replaces characters that are invalid in filenames
func sanitizeFilename(name string) string {
	// Replace spaces with underscores
	name = strings.ReplaceAll(name, " ", "_")

	// Remove or replace problematic characters
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "",
		"?", "",
		"\"", "",
		"<", "",
		">", "",
		"|", "",
		".", "",
	)
	name = replacer.Replace(name)

	// Limit length to avoid filesystem issues
	if len(name) > 50 {
		name = name[:50]
	}

	return name
}

// generateHTMLViewer creates a self-contained HTML viewer with embedded data
func generateHTMLViewer(outputDir string, treeExport *TreeExport) error {
	// Read the people.json file we just created (it has relationships + media embedded)
	peopleJSONPath := filepath.Join(outputDir, "people.json")
	peopleJSON, err := os.ReadFile(peopleJSONPath)
	if err != nil {
		return fmt.Errorf("failed to read people.json: %w", err)
	}

	// Marshal metadata
	metadata := map[string]interface{}{
		"treeId":      treeExport.TreeID,
		"treeName":    treeExport.TreeName,
		"exportDate":  treeExport.ExportDate,
		"personCount": treeExport.PersonCount,
	}
	metadataJSON, _ := json.Marshal(metadata)

	// Generate main index HTML with embedded data
	htmlContent := generateHTMLTemplate(string(peopleJSON), string(metadataJSON))
	htmlPath := filepath.Join(outputDir, "index.html")
	if err := os.WriteFile(htmlPath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("failed to write index.html: %w", err)
	}

	// Generate single person page that uses URL parameters
	personHTML := generatePersonPageTemplate(string(peopleJSON), string(metadataJSON))
	personPath := filepath.Join(outputDir, "person.html")
	if err := os.WriteFile(personPath, []byte(personHTML), 0644); err != nil {
		return fmt.Errorf("failed to write person.html: %w", err)
	}

	return nil
}
