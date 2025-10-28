package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/urfave/cli/v3"
)

const (
	ShortUsage      = "A web server for prototyping web interfaces from sketched images."
	LongDescription = `The server uses a config.yaml file to describe clickable zones on
   images in assets/images to build an interactive website.
   
   For a demo with embedded assets and config file, use 'demo'.
   To start a new project, use 'init' to write the demo files to disk.
   To serve files on disk use 'serve'.`
)

// Applicator is an interface to the central coordinator for the project
// (concretely provided by App in app.go) to allow for testing.
type Applicator interface {
	Serve(address, port, configFile string) error
	Init(directory string) error
	Demo(address, port string) error
}

// BuildCLI creates a cli app to run the capabilities provided by
// an Applicator dependency.
func BuildCLI(app Applicator) *cli.Command {

	// Define the common flags.
	addressFlag := &cli.StringFlag{
		Name:    "address",
		Aliases: []string{"a"},
		Value:   "127.0.0.1",
		Usage:   "server network address",
	}
	portFlag := &cli.StringFlag{
		Name:    "port",
		Aliases: []string{"p"},
		Value:   "8000",
		Usage:   "server network port",
	}

	serveCmd := &cli.Command{
		Name:      "serve",
		Usage:     "Serve content on disk with the provided config",
		ArgsUsage: "CONFIG_FILE",
		// use the common flags
		Flags: []cli.Flag{
			addressFlag,
			portFlag,
		},
		// Before runs verification before "Action" is run
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			if c.NArg() < 1 {
				return ctx, fmt.Errorf("missing required argument: CONFIG_FILE")
			}
			configFile := c.Args().First()
			if _, err := os.Stat(configFile); err != nil {
				return ctx, fmt.Errorf("config file %q not found", configFile)
			}
			if a := net.ParseIP(c.String("address")); a == nil {
				return ctx, fmt.Errorf("invalid IP address: %s", c.String("address"))
			}
			if _, err := strconv.Atoi(c.String("port")); err != nil {
				return ctx, fmt.Errorf("invalid port: %s", c.String("port"))
			}
			return ctx, nil
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			configFile := c.Args().First()
			return app.Serve(c.String("address"), c.String("port"), configFile)
		},
	}

	initCmd := &cli.Command{
		Name:  "init",
		Usage: "Initialize a new project in a directory",
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			dir := c.String("directory")
			d, err := os.Stat(dir)
			if err != nil {
				return ctx, fmt.Errorf("directory %q does not exist", dir)
			}
			if !d.IsDir() {
				return ctx, fmt.Errorf("%q is not a directory", dir)
			}
			return ctx, nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "directory",
				Aliases: []string{"d"},
				Value:   ".", // better than os.Getwd
				Usage:   "directory to write files",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			return app.Init(c.String("directory"))
		},
	}

	demoCmd := &cli.Command{
		Name:                  "demo",
		Usage:                 "Run the embedded demo server",
		EnableShellCompletion: true,
		// use the common flags
		Flags: []cli.Flag{
			addressFlag,
			portFlag,
		},
		// Repeat validation logic (consider sharing).
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			if a := net.ParseIP(c.String("address")); a == nil {
				return ctx, fmt.Errorf("invalid IP address: %s", c.String("address"))
			}
			if _, err := strconv.Atoi(c.String("port")); err != nil {
				return ctx, fmt.Errorf("invalid port: %s", c.String("port"))
			}
			return ctx, nil
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			return app.Demo(c.String("address"), c.String("port"))
		},
	}

	rootCmd := &cli.Command{
		Name:        "firstgo",
		Usage:       ShortUsage,
		Description: LongDescription,
		Commands:    []*cli.Command{serveCmd, initCmd, demoCmd},
	}

	// custom help template.
	rootCmd.CustomRootCommandHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   {{.Name}} [global options] [command]

DESCRIPTION:
   {{.Description}}

COMMANDS:
{{range .Commands}}   {{.Name}}{{ "\t"}}{{.Usage}}
{{end}}
Run '{{.Name}} [command] --help' for more information on a command.
`

	return rootCmd
}
