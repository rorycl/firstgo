package main

import (
	"errors"
	"os"
	"testing"
)

func TestApp(t *testing.T) {

	ServeTestOK := func(*server) error {
		return nil
	}
	ServeTestFail := func(*server) error {
		return errors.New("fail")
	}
	WriteTestOK := func(cfg *config, directory string) error {
		return nil
	}
	WriteTestFail := func(cfg *config, directory string) error {
		return errors.New("fail")
	}

	appOK := NewApp()
	appOK.serveFunc = ServeTestOK
	appOK.writeFunc = WriteTestOK

	appFail := NewApp()
	appFail.serveFunc = ServeTestFail
	appFail.writeFunc = WriteTestFail

	// pass cases
	err := appOK.Serve("127.0.0.1", "8000", "config.yaml")
	if err != nil {
		t.Fatalf("got unexepected error %v", err)
	}
	err = appOK.Demo("127.0.0.1", "8000")
	if err != nil {
		t.Fatalf("got unexepected error %v", err)
	}
	err = appOK.Init(os.TempDir())
	if err != nil {
		t.Fatalf("got unexepected error %v", err)
	}

	// error cases
	err = appFail.Serve("127.0.0.1", "8000", "config.yaml")
	if err == nil {
		t.Fatal("unexpected success for app failure mode")
	}
	err = appFail.Demo("127.0.0.1", "8000")
	if err == nil {
		t.Fatal("unexpected success for app failure mode")
	}
	err = appFail.Init(os.TempDir())
	if err == nil {
		t.Fatal("unexpected success for app failure mode")
	}

	// interactive
	if got, want := appOK.interactive, false; got != want {
		t.Errorf("interactive status got %t want %t", got, want)
	}
	appOK.Interactive()
	if got, want := appOK.interactive, true; got != want {
		t.Errorf("interactive status after switch got %t want %t", got, want)
	}
	appOK.Interactive()
	if got, want := appOK.interactive, false; got != want {
		t.Errorf("interactive status after second switch got %t want %t", got, want)
	}

}
