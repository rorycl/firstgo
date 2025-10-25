package main

import (
	"fmt"
	"os"
)

func main() {
	app := NewApp()
	msg, err := ParseFlags(app)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	if msg != "" {
		fmt.Println(msg)
	}
}
