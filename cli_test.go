package main

import (
	"context"
	"io"
	"strings"
	"testing"
)

// TestApplication implements the Applicator interface.
type TestApplication struct{}

func (t *TestApplication) Serve(address, port, configFile string) error {
	return nil
}
func (t *TestApplication) ServeInDevelopment(address, port string, templateSuffixes []string, configFile string) error {
	return nil
}
func (t *TestApplication) Init(directory string) error {
	return nil
}
func (t *TestApplication) Demo(address, port string) error {
	return nil
}

func TestParseCLI(t *testing.T) {

	testApp := &TestApplication{}

	tests := []struct {
		name            string
		args            []string
		wantErrContains string
	}{
		{
			name: "help",
			args: []string{"program", "-h"},
		},
		{
			name: "serve all options",
			args: []string{"program", "serve", "-a", "127.0.0.1", "-p", "8001", "config.yaml"},
		},
		{
			name: "serve only config",
			args: []string{"program", "serve", "config.yaml"},
		},
		{
			name:            "serve no config",
			args:            []string{"program", "serve", "--address", "127.0.0.2"},
			wantErrContains: "missing required argument",
		},
		{
			name: "serve help",
			args: []string{"program", "serve", "-h"},
		},
		{
			name:            "serve invalid address",
			args:            []string{"program", "serve", "-a", "hi", "-p", "8001", "config.yaml"},
			wantErrContains: "invalid IP address",
		},
		{
			name:            "serve invalid port",
			args:            []string{"program", "serve", "-a", "127.0.0.3", "-p", "eight", "config.yaml"},
			wantErrContains: "invalid port",
		},
		{
			name: "init help",
			args: []string{"program", "init", "-h"},
		},
		{
			name: "init ok default",
			args: []string{"program", "init"},
		},
		{
			name: "init ok with tmp dir",
			args: []string{"program", "init", "-d", "/tmp"},
		},
		{
			name:            "init failure",
			args:            []string{"program", "init", "-d", "/_DATA/tmp"},
			wantErrContains: "does not exist",
		},
		{
			name: "demo help",
			args: []string{"program", "demo", "-h"},
		},
		{
			name: "demo ok",
			args: []string{"program", "demo", "-a", "127.0.0.1", "-p", "8001"},
		},
		{
			name: "demo ok no args",
			args: []string{"program", "demo"},
		},
		{
			name:            "demo invalid address",
			args:            []string{"program", "demo", "-a", "url", "-p", "8001"},
			wantErrContains: "invalid IP address",
		},
		{
			name: "development all options",
			args: []string{"program", "development", "-a", "127.0.0.1", "-p", "8001", "-s", "html", "config.yaml"},
		},
		{
			name: "development only config",
			args: []string{"program", "development", "config.yaml"},
		},
		{
			name:            "development no config",
			args:            []string{"program", "development", "--address", "127.0.0.2"},
			wantErrContains: "missing required argument",
		},
		{
			name:            "development no suffix",
			args:            []string{"program", "development", "-a", "127.0.0.1", "-p", "8001", "-s", "", "config.yaml"},
			wantErrContains: "empty suffix argument",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := BuildCLI(testApp)
			cmd.Writer = io.Discard
			cmd.ErrWriter = io.Discard
			err := cmd.Run(context.Background(), tt.args)
			if tt.wantErrContains != "" {
				if err == nil {
					t.Fatalf("expected an error containing %q", tt.wantErrContains)
				}
				if got, want := err.Error(), tt.wantErrContains; !strings.Contains(got, want) {
					t.Fatalf("got error that did not contain %q: %v", want, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
