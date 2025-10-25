package main

import (
	"fmt"
	"os"
)

func main() {

	app := NewApp()
	err := ParseFlags(app)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
