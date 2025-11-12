package commands

import (
	"fmt"
	"time"

	"github.com/chrisrob11/ancestrydl/pkg/ancestry"
	"github.com/urfave/cli/v2"
)

// getTreeOwnerID retrieves the owner ID, checking both possible fields
func getTreeOwnerID(tree ancestry.Tree) string {
	if tree.Owner != "" {
		return tree.Owner
	}
	return tree.OwnerUserID
}

// getTreeCreatedDate retrieves the creation date, checking both possible fields
func getTreeCreatedDate(tree ancestry.Tree) time.Time {
	if !tree.CreatedOn.IsZero() {
		return tree.CreatedOn
	}
	return tree.DateCreated
}

// getTreeModifiedDate retrieves the modification date, checking both possible fields
func getTreeModifiedDate(tree ancestry.Tree) time.Time {
	if !tree.ModifiedOn.IsZero() {
		return tree.ModifiedOn
	}
	return tree.DateModified
}

// displayTreeInfo prints formatted information for a single tree
func displayTreeInfo(i int, tree ancestry.Tree) {
	fmt.Printf("[%d] %s\n", i+1, tree.Name)
	fmt.Printf("    ID: %s\n", tree.ID)

	if ownerID := getTreeOwnerID(tree); ownerID != "" {
		fmt.Printf("    Owner: %s\n", ownerID)
	}

	if createdDate := getTreeCreatedDate(tree); !createdDate.IsZero() {
		fmt.Printf("    Created: %s\n", createdDate.Format("2006-01-02"))
	}

	if modifiedDate := getTreeModifiedDate(tree); !modifiedDate.IsZero() {
		fmt.Printf("    Modified: %s\n", modifiedDate.Format("2006-01-02"))
	}

	if tree.Description != "" {
		fmt.Printf("    Description: %s\n", tree.Description)
	}

	if tree.CanSeeLiving {
		fmt.Printf("    Can See Living: Yes\n")
	}

	if tree.SH {
		fmt.Printf("    Shared: Yes\n")
	}

	if tree.TotalInvitedCount > 0 {
		fmt.Printf("    Total Invited: %d\n", tree.TotalInvitedCount)
	}

	fmt.Println()
}

// ListTrees retrieves and displays all family trees for the authenticated user
func ListTrees(c *cli.Context) error {
	fmt.Println("Retrieving your family trees...")
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

	fmt.Println("Fetching trees from Ancestry.com...")
	trees, err := apiClient.ListTrees()
	if err != nil {
		return fmt.Errorf("failed to retrieve trees: %w\n\nYour session may have expired. Try running 'ancestrydl login' again", err)
	}

	fmt.Println()
	if len(trees) == 0 {
		fmt.Println("No trees found.")
		return nil
	}

	fmt.Printf("Found %d tree(s):\n\n", len(trees))

	for i, tree := range trees {
		displayTreeInfo(i, tree)
	}

	return nil
}
