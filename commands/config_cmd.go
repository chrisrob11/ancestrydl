package commands

import (
	"fmt"

	"github.com/chrisrob11/ancestrydl/pkg/config"
	"github.com/urfave/cli/v2"
)

// SetDefaultTree sets the default tree ID in config
func SetDefaultTree(c *cli.Context) error {
	treeID := c.Args().First()
	if treeID == "" {
		return fmt.Errorf("tree ID is required\n\nUsage: ancestrydl config set-default-tree <tree-id>")
	}

	if err := config.SetDefaultTreeID(treeID); err != nil {
		return fmt.Errorf("failed to set default tree: %w", err)
	}

	fmt.Printf("✓ Default tree set to: %s\n", treeID)
	fmt.Println()
	fmt.Println("You can now run commands without specifying a tree ID:")
	fmt.Println("  ancestrydl list-people")
	fmt.Println()

	return nil
}

// ShowConfig displays the current configuration
func ShowConfig(c *cli.Context) error {
	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Current configuration:")
	fmt.Println()

	if cfg.DefaultTreeID != "" {
		fmt.Printf("  Default Tree ID: %s\n", cfg.DefaultTreeID)
	} else {
		fmt.Println("  Default Tree ID: (not set)")
	}

	fmt.Println()
	fmt.Println("Config file: ~/.ancestrydl/config.json")

	return nil
}
