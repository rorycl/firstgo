package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"testing"
)

func TestMainFlags(t *testing.T) {

	tests := []struct {
		args       []string
		address    string
		port       string
		configFile string
		err        error
	}{
		{
			args:       []string{"prog", "-address", "10.0.1.1", "-port", "8021", "config.yaml"},
			address:    "10.0.1.1",
			port:       "8021",
			configFile: "config.yaml",
			err:        nil,
		},
		{
			args:       []string{"prog", "config.yaml"},
			address:    "127.0.0.1",
			port:       "8000",
			configFile: "config.yaml",
			err:        nil,
		},
		{
			args: []string{"prog"},
			err:  ErrFlagExited,
		},
		{
			args: []string{"prog", "-address", "10.2000.1.1", "config.yaml"},
			err:  errors.New("invalid ip"),
		},
		{
			args: []string{"prog", "-port", "hi", "config.yaml"},
			err:  errors.New("invalid port"),
		},
		{
			args: []string{"prog", "no config.yaml"},
			err:  errors.New("no config file found"),
		},
	}

	for ii, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", ii), func(t *testing.T) {

			flag.CommandLine = flag.NewFlagSet(fmt.Sprintf("%d", ii), flag.ContinueOnError)
			os.Args = tt.args
			address, port, configFile, err := flagGet()

			if tt.err == nil && err != nil {
				t.Fatalf("unexpected err %v", err)
			}
			if tt.err != nil && errors.Is(err, tt.err) {
				return
			}
			if tt.err != nil && err != nil {
				// assume ok
				t.Logf("got err %v (expected %v)", err, tt.err)
				return
			}
			if got, want := address, tt.address; got != want {
				t.Errorf("address got %v want %v", got, want)
			}
			if got, want := port, tt.port; got != want {
				t.Errorf("port got %v want %v", got, want)
			}
			if got, want := configFile, tt.configFile; got != want {
				t.Errorf("configFile got %v want %v", got, want)
			}
		})
	}
}

func TestMainMain(t *testing.T) {

	var exitCode int
	Exiter = func(e int) {
		exitCode = e
	}

	Serve = func(*server) error { return nil }
	flag.CommandLine = flag.NewFlagSet("test-main", flag.ContinueOnError)
	os.Args = []string{"prog", "config.yaml"}
	address, port, configFile, err := flagGet()
	if err != nil {
		t.Fatalf("flag parse error at step %s, %v", "one", err)
	}

	configBytes, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatal(err)
	}

	config, err := newConfig(configBytes)
	if err != nil {
		t.Fatalf("config file error at step %s, %v", "one", err)
	}

	server, err := newServer(address, port, config.Pages, config.PageTemplate)
	if err != nil {
		t.Fatalf("new server error at step %s, %v", "one", err)
	}

	err = Serve(server)
	if err != nil || exitCode != 0 {
		t.Fatalf("serve error at step %s, %v : code %d", "one", err, exitCode)
	}

}
