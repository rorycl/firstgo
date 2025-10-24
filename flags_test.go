package main

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/jessevdk/go-flags"
)

func TestParseFlags(t *testing.T) {

	tests := []struct {
		name string
		args []string
		err  error
	}{
		{
			name: "serve: all options",
			args: []string{"program", "serve", "-a", "127.0.0.1", "-p", "8001", "config.yaml"},
			err:  nil,
		},
		{
			name: "serve: only config",
			args: []string{"program", "serve", "config.yaml"},
			err:  nil,
		},
		{
			name: "serve: no config",
			args: []string{"program", "serve", "--address", "127.0.0.2"},
			err:  &flags.Error{},
		},
		{
			name: "serve: invalid address",
			args: []string{"program", "serve", "-a", "hi", "-p", "8001", "config.yaml"},
			err:  FlagCustomError{},
		},
		{
			name: "serve: invalid port",
			args: []string{"program", "serve", "-a", "127.0.0.3", "-p", "eight", "config.yaml"},
			err:  FlagCustomError{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			_, err := ParseFlags()
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
