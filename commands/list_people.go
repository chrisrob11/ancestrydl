package commands

import (
	"fmt"

	"github.com/chrisrob11/ancestrydl/pkg/ancestry"
	"github.com/chrisrob11/ancestrydl/pkg/config"
	"github.com/urfave/cli/v2"
)

// getTreeIDOrDefault retrieves the tree ID from arguments or uses the default
func getTreeIDOrDefault(c *cli.Context) (string, error) {
	treeID := c.Args().First()
	if treeID != "" {
		return treeID, nil
	}

	defaultTreeID, err := config.GetDefaultTreeID()
	if err != nil {
		return "", fmt.Errorf("failed to get default tree: %w", err)
	}
	if defaultTreeID == "" {
		return "", fmt.Errorf("tree ID is required\n\nUsage: ancestrydl list-people <tree-id>\n\nOr set a default tree with: ancestrydl config set-default-tree <tree-id>")
	}
	fmt.Printf("Using default tree: %s\n", defaultTreeID)
	return defaultTreeID, nil
}

// createAPIClientFromStoredCookies creates an API client from stored session cookies
func createAPIClientFromStoredCookies() (*ancestry.APIClient, error) {
	cookiesJSON, err := config.GetCookies()
	if err != nil {
		return nil, fmt.Errorf("failed to load stored cookies: %w\n\nPlease run 'ancestrydl login' first to authenticate", err)
	}
	if cookiesJSON == "" {
		return nil, fmt.Errorf("no stored cookies found\n\nPlease run 'ancestrydl login' first to authenticate")
	}

	apiClient, err := ancestry.NewAPIClientFromJSON(cookiesJSON, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}
	return apiClient, nil
}

// fetchAllPersons retrieves all persons from a tree with pagination
func fetchAllPersons(apiClient *ancestry.APIClient, treeID string, totalCount int) ([]ancestry.Person, error) {
	limit := 100
	totalPages := (totalCount + limit - 1) / limit

	fmt.Printf("Fetching %d page(s) of data...\n", totalPages)
	fmt.Println()

	allPersons := []ancestry.Person{}
	for page := 1; page <= totalPages; page++ {
		fmt.Printf("Fetching page %d/%d...\n", page, totalPages)
		persons, err := apiClient.GetAllPersons(treeID, page, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve persons on page %d: %w", page, err)
		}
		allPersons = append(allPersons, persons...)
	}
	return allPersons, nil
}

// getPersonName extracts the display name from a person
func getPersonName(person ancestry.Person) string {
	if person.GivenName != "" || person.Surname != "" {
		return fmt.Sprintf("%s %s", person.GivenName, person.Surname)
	}
	if len(person.Names) > 0 {
		return fmt.Sprintf("%s %s", person.Names[0].GivenName, person.Names[0].Surname)
	}
	return ""
}

// getPersonLifeEvents extracts birth and death years from events
func getPersonLifeEvents(person ancestry.Person) (birthYear, deathYear string) {
	for _, event := range person.Events {
		if event.Type == "Birth" && event.Date != nil {
			birthYear = fmt.Sprintf("%v", event.Date)
		}
		if event.Type == "Death" && event.Date != nil {
			deathYear = fmt.Sprintf("%v", event.Date)
		}
	}
	return birthYear, deathYear
}

// displayPerson prints formatted person information
func displayPerson(i int, person ancestry.Person) {
	name := getPersonName(person)
	birthYear, deathYear := getPersonLifeEvents(person)

	fmt.Printf("[%d] %s\n", i+1, name)
	if personID := person.GetPersonID(); personID != "" {
		fmt.Printf("    ID: %s\n", personID)
	}
	if person.Gender != "" {
		fmt.Printf("    Gender: %s\n", person.Gender)
	}
	if birthYear != "" {
		fmt.Printf("    Birth: %s\n", birthYear)
	}
	if deathYear != "" {
		fmt.Printf("    Death: %s\n", deathYear)
	}
	if person.IsLiving {
		fmt.Printf("    Living: Yes\n")
	}
	fmt.Println()
}

// ListPeople retrieves and displays all people in a family tree
func ListPeople(c *cli.Context) error {
	treeID, err := getTreeIDOrDefault(c)
	if err != nil {
		return err
	}

	fmt.Printf("Retrieving people from tree %s...\n", treeID)
	fmt.Println()

	fmt.Println("Creating API client from stored session...")
	apiClient, err := createAPIClientFromStoredCookies()
	if err != nil {
		return err
	}
	defer func() {
		if err := apiClient.Close(); err != nil {
			fmt.Printf("Error closing API client: %v\n", err)
		}
	}()

	fmt.Println("Getting person count...")
	totalCount, err := apiClient.GetPersonsCount(treeID)
	if err != nil {
		return fmt.Errorf("failed to get person count: %w\n\nYour session may have expired. Try running 'ancestrydl login' again", err)
	}

	fmt.Printf("Tree has %d total persons\n", totalCount)
	fmt.Println()

	if totalCount == 0 {
		fmt.Println("No people found in this tree.")
		return nil
	}

	allPersons, err := fetchAllPersons(apiClient, treeID, totalCount)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("Successfully retrieved %d person(s):\n\n", len(allPersons))

	for i, person := range allPersons {
		displayPerson(i, person)
	}

	return nil
}
