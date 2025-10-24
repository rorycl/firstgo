package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/jessevdk/go-flags"
)

// FlagCustomError is a custom flag parsing error.
type FlagCustomError struct {
	message string
}

// Error allows FlagCustomError to act as an error.
func (f FlagCustomError) Error() string {
	return f.message
}

// Options is the global flag options wrapper struct.
type Options struct {
	// can we add "serve command" etc in here?
}

// ServeCommand are the flag options for the '<prog> serve' command.
type ServeCommand struct {
	Address string `short:"a" long:"address" description:"server address" default:"127.0.0.1"`
	Port    string `short:"p" long:"port" description:"server port" default:"8000"`
	Args    struct {
		ConfigFile string `description:"configuration yaml file"`
	} `positional-args:"yes" required:"yes"`
	// injected main app
	App *App `no-flag:"true"`
}

// validate runs additional checks on the ServeCommand.
func (c *ServeCommand) validate() error {
	if a := net.ParseIP(c.Address); a == nil {
		return FlagCustomError{
			fmt.Sprintf("address %s is an invalid IP address", c.Address),
		}
	}
	if _, err := strconv.Atoi(c.Port); err != nil {
		return FlagCustomError{
			fmt.Sprintf("port %s invalid", c.Port),
		}
	}
	return nil
}

// Execute runs the server.
func (c *ServeCommand) Execute(args []string) error {
	if err := c.validate(); err != nil {
		return err
	}
	return c.App.Serve(c.Address, c.Port, c.Args.ConfigFile)
}

// InitCommand are the flag options for the '<prog> init' command.
type InitCommand struct {
	Directory string `short:"d" long:"directory" description:"directory to write files" default:"cwd"`
	// injected main app
	App *App `no-flag:"true"`
}

// Execute runs the Init process.
func (c *InitCommand) Execute(args []string) error {
	var err error
	if c.Directory == "cwd" {
		c.Directory, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("could not get current working directory %v", err)
		}
	}
	return c.App.Init(c.Directory)
}

// DemoCommand are the flag options for the '<prog> demo' command.
type DemoCommand struct {
	Address string `short:"a" long:"address" description:"server address" default:"127.0.0.1"`
	Port    string `short:"p" long:"port" description:"server port" default:"8000"`
	// injected main app
	App *App `no-flag:"true"`
}

// Execute runs the Demo process.
func (c *DemoCommand) Execute(args []string) error {
	return c.App.Demo(c.Address, c.Port)
}

var cmdTpl string = `

A web server for prototyping web interfaces using sketches and clickable
zones to move between pages.

Modes:

	demo  : show a demo site with embedded assets
	init  : materialize the demo site to disk to init a project
	serve : serve a site from disk
`

// ParseFlags parses the command line options.
func ParseFlags() (*Options, error) {
	var options Options
	var parser = flags.NewParser(&options, flags.HelpFlag)
	parser.Usage = cmdTpl

	// Add the 'serve' command
	parser.AddCommand(
		"serve",
		"Serve content on disk with the provided config",
		"The serve command starts the web server with the provided yaml configuration file.",
		&ServeCommand{},
	)

	// Add the 'init' command
	parser.AddCommand(
		"init",
		"Init a project",
		"Initialise a project by writing the embedded demo project to disk.",
		&InitCommand{},
	)

	// Add the 'demo' command
	parser.AddCommand(
		"demo",
		"Run the demo server",
		"Serve the demo content embedded in the program.",
		&ServeCommand{},
	)

	// Catch errors in caller.
	if _, err := parser.Parse(); err != nil {
		return nil, err
	}

	fmt.Printf("%#v\n", options)

	return &options, nil
}
