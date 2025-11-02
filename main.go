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
				Name:    "export",
				Aliases: []string{"e"},
				Usage:   "Export a family tree with all media",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tree-id",
						Aliases:  []string{"t"},
						Usage:    "Ancestry tree ID",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "Output directory",
						Value:   "./ancestry-export",
					},
				},
				Action: exportCommand,
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
	return fmt.Errorf("list-trees command not yet implemented")
}

func exportCommand(c *cli.Context) error {
	return fmt.Errorf("export command not yet implemented")
}
