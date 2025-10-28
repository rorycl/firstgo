package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	// init core app capabilities with console messages on.
	app := NewApp()
	app.Interactive()

	// build cli, injecting app
	cmd := BuildCLI(app)

	// run
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
