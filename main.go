package main

import (
	"fmt"
	"log"
	"os"

	"github.com/chrisrob11/ancestrydl/commands"
	"github.com/urfave/cli/v2"
)

var (
	// Version is set via ldflags during build
	Version = "dev"
	// BuildDate is set via ldflags during build
	BuildDate = "unknown"
)

func main() {
	app := &cli.App{
		Name:    "ancestrydl",
		Usage:   "Download your family tree data from Ancestry.com",
		Version: fmt.Sprintf("%s (built %s)", Version, BuildDate),
		Commands: []*cli.Command{
			{
				Name:    "login",
				Aliases: []string{"l"},
				Usage:   "Authenticate with Ancestry.com",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "username",
						Aliases:  []string{"u"},
						Usage:    "Ancestry.com email/username",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "password",
						Aliases:  []string{"p"},
						Usage:    "Ancestry.com password",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "2fa",
						Usage: "2FA method to auto-select: 'email' or 'phone' (if account has 2FA enabled)",
					},
				},
				Action: loginCommand,
			},
			{
				Name:    "logout",
				Aliases: []string{"lo"},
				Usage:   "Remove stored credentials",
				Action:  logoutCommand,
			},
			{
				Name:    "list-trees",
				Aliases: []string{"ls"},
				Usage:   "List all available family trees",
				Action:  listTreesCommand,
			},
			{
				Name:      "list-people",
				Aliases:   []string{"lp"},
				Usage:     "List all people in a family tree",
				ArgsUsage: "<tree-id>",
				Action:    listPeopleCommand,
			},
			{
				Name:    "config",
				Aliases: []string{"cfg"},
				Usage:   "Manage configuration settings",
				Subcommands: []*cli.Command{
					{
						Name:      "set-default-tree",
						Usage:     "Set the default tree ID",
						ArgsUsage: "<tree-id>",
						Action:    setDefaultTreeCommand,
					},
					{
						Name:    "show",
						Aliases: []string{"s"},
						Usage:   "Show current configuration",
						Action:  showConfigCommand,
					},
				},
			},
			{
				Name:      "download-tree",
				Aliases:   []string{"dl"},
				Usage:     "Download complete family tree with all data and media",
				ArgsUsage: "[tree-id]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "Output directory",
						Value:   "./ancestry-export",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Enable verbose logging (writes all HTTP requests/responses to http_log.txt)",
					},
				},
				Action: downloadTreeCommand,
			},
			{
				Name:  "test-browser",
				Usage: "Test browser automation (opens browser and navigates to Ancestry.com)",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "username",
						Aliases: []string{"u"},
						Usage:   "Ancestry.com email/username (optional, for testing login)",
					},
					&cli.StringFlag{
						Name:    "password",
						Aliases: []string{"p"},
						Usage:   "Ancestry.com password (optional, for testing login)",
					},
					&cli.StringFlag{
						Name:  "2fa",
						Usage: "2FA method to auto-select: 'email' or 'phone' (optional)",
					},
					&cli.BoolFlag{
						Name:  "no-submit",
						Usage: "Fill login form but don't submit (for testing/debugging)",
					},
					&cli.BoolFlag{
						Name:  "capture-network",
						Usage: "Capture and display all API network requests",
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "Save captured network requests to file (JSON format)",
					},
				},
				Action: testBrowserCommand,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

// Command stubs - to be implemented in separate commits
func loginCommand(c *cli.Context) error {
	return commands.Login(c)
}

func logoutCommand(c *cli.Context) error {
	return commands.Logout(c)
}

func listTreesCommand(c *cli.Context) error {
	return commands.ListTrees(c)
}

func listPeopleCommand(c *cli.Context) error {
	return commands.ListPeople(c)
}

func setDefaultTreeCommand(c *cli.Context) error {
	return commands.SetDefaultTree(c)
}

func showConfigCommand(c *cli.Context) error {
	return commands.ShowConfig(c)
}

func downloadTreeCommand(c *cli.Context) error {
	return commands.DownloadTree(c)
}

func testBrowserCommand(c *cli.Context) error {
	return commands.TestBrowser(c)
}
