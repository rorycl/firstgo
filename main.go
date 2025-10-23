package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

var usage = `
run a "firstgo" webserver to show interactive image "pages".

This programme uses the configuration in config.yaml and images in the
images directory and css in the static directory to serve up interactive
image "web pages" to mock up a web site or service.

eg ./firstgo [-address 192.168.4.5] [-port 8001] <configfile>
`

var ErrFlagExited error = errors.New("an error occurred")

// flagGet checks the flags
func flagGet() (string, string, string, error) {

	var address, port, configFile string
	flag.StringVar(&address, "address", "127.0.0.1", "server network address")
	flag.StringVar(&port, "port", "8000", "server network port")

	flag.Usage = func() {
		// supressing Fprint errors
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		_, _ = fmt.Fprintln(
			flag.CommandLine.Output(),
			"  <configFile>\n    	yaml configuration file")
		_, _ = fmt.Fprint(flag.CommandLine.Output(), usage)
	}

	flag.Parse()
	if address == "" || port == "" {
		flag.Usage()
		return "", "", "", ErrFlagExited
	}
	if len(flag.Args()) != 1 {
		flag.Usage()
		return "", "", "", ErrFlagExited
	}
	configFile = flag.Args()[0]

	// check validity of fields
	a := net.ParseIP(address)
	if a == nil {
		return "", "", "", fmt.Errorf("address %s invalid IP address", address)
	}
	_, err := strconv.Atoi(port)
	if err != nil {
		return "", "", "", fmt.Errorf("port %s invalid", port)
	}
	if _, err = os.Stat(configFile); err != nil {
		return "", "", "", fmt.Errorf("config file %s could not be found", configFile)
	}
	return address, port, configFile, nil
}

// Serve indirects serve for testing
var Serve func(*server) error = (*server).serve

var Exiter func(int) = os.Exit

func main() {

	address, port, configFile, err := flagGet()
	if err != nil {
		if !errors.Is(err, ErrFlagExited) {
			fmt.Println(err)
		}
		Exiter(1)
		return
	}

	configBytes, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Println(err)
		Exiter(1)
		return
	}

	config, err := newConfig(configBytes)
	if err != nil {
		fmt.Println(err)
		Exiter(1)
		return
	}

	server, err := newServer(address, port, config.Pages, config.PageTemplate)
	if err != nil {
		fmt.Println(err)
		Exiter(1)
		return
	}

	log.Printf("serving on %s:%s", address, port)
	err = Serve(server)
	// err = server.serve()
	if err != nil {
		fmt.Println(err)
		Exiter(1)
		return
	}
	Exiter(0)
}
