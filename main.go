package main

import (
	"fmt"
	"os"
)

func main() {

	config, err := newConfig("config.yaml")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	server, err := newServer(config.Pages, config.PageTemplate)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = server.serve()
	if err != nil {
		fmt.Println(err)
	}

}
