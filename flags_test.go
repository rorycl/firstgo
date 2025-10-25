package main

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/jessevdk/go-flags"
)

// TestApplication implements the Applicator interface.
type TestApplication struct{}

func (t *TestApplication) Serve(address, port, configFile string) error {
	return nil
}
func (t *TestApplication) Init(directory string) error {
	return nil
}
func (t *TestApplication) Demo(address, port string) error {
	return nil
}

func TestParseFlags(t *testing.T) {

	testApp := &TestApplication{}

	tests := []struct {
		name string
		args []string
		err  error
	}{
		{
			name: "serve all options",
			args: []string{"program", "serve", "-a", "127.0.0.1", "-p", "8001", "config.yaml"},
			err:  nil,
		},
		{
			name: "serve only config",
			args: []string{"program", "serve", "config.yaml"},
			err:  nil,
		},
		{
			name: "serve no config",
			args: []string{"program", "serve", "--address", "127.0.0.2"},
			err:  &flags.Error{},
		},
		{
			name: "serve invalid address",
			args: []string{"program", "serve", "-a", "hi", "-p", "8001", "config.yaml"},
			err:  FlagCustomError{},
		},
		{
			name: "serve invalid port",
			args: []string{"program", "serve", "-a", "127.0.0.3", "-p", "eight", "config.yaml"},
			err:  FlagCustomError{},
		},
		{
			name: "init ok default",
			args: []string{"program", "init"},
			err:  nil,
		},
		{
			name: "init ok with tmp dir",
			args: []string{"program", "init", "-d", "/tmp"},
			err:  nil,
		},
		{
			name: "init failure",
			args: []string{"program", "init", "-d", "/_DATA/tmp"},
			err:  FlagCustomError{},
		},
		{
			name: "demo ok",
			args: []string{"program", "demo", "-a", "127.0.0.1", "-p", "8001"},
			err:  nil,
		},
		{
			name: "demo ok no args",
			args: []string{"program", "demo"},
			err:  nil,
		},
		{
			name: "demo invalid address",
			args: []string{"program", "demo", "-a", "url", "-p", "8001"},
			err:  FlagCustomError{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			err := ParseFlags(testApp)
			if tt.err != nil {
				if err == nil {
					t.Fatal("expected an error")
				}
				if errors.Is(err, tt.err) && fmt.Sprintf("%T", tt.err) != "*errors.errorString" {
					t.Fatalf("got error %q expected type %T", err, tt.err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestFlagCustomError(t *testing.T) {
	e := FlagCustomError{"hi"}
	var fce FlagCustomError
	if !errors.As(e, &fce) {
		t.Fatal("expected FlagCustomError")
	}
	if got, want := e.Error(), "hi"; got != want {
		t.Errorf("error got %s want %s", got, want)
	}
}
