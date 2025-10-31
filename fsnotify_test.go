package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var dir1, dir2 string

func mkdirs(t *testing.T) {
	t.Helper()
	var err error
	dir1, err = os.MkdirTemp("", "firstgo_dir*")
	if err != nil {
		t.Fatal(err)
	}
	dir2, err = os.MkdirTemp("", "firstgo_dir*")
	if err != nil {
		t.Fatal(err)
	}
}

func writeFiles(t *testing.T) {
	t.Helper()
	for _, dirFile := range [][]string{
		[]string{dir1, ".newfile.html"}, // not counted
		[]string{dir1, "abc.html"},      // counted
		[]string{dir1, "abc.HTML"},      // counted
		[]string{dir1, ".hidden.HTML"},  // not counted
		[]string{dir2, "abctxt"},        // not counted
		[]string{dir2, "ABC.txt"},       // counted
	} {
		dir, file := dirFile[0], dirFile[1]
		fmt.Println(dir, file)
		o, err := os.Create(filepath.Join(dir, file))
		if err != nil {
			t.Fatal(err)
		}
		_, err = fmt.Fprint(o, "hi")
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(flushDuration + 2) // accommodate flush interval
	}
}

func TestFileChangeNotifierRefresh(t *testing.T) {

	mkdirs(t)

	t.Cleanup(func() {
		_ = os.RemoveAll(dir1)
		_ = os.RemoveAll(dir2)
	})

	ctx, cancel := context.WithCancel(context.Background())
	fcn, err := NewFileChangeNotifier(
		ctx,
		[]DirFilesDescriptor{
			DirFilesDescriptor{dir1, []string{".html"}},
			DirFilesDescriptor{dir2, []string{"txt"}},
		},
	)
	if err != nil {
		t.Fatalf("error initialising fcn: %v", err)
	}

	counter := 0
	go func() {
		for _, err = range fcn.Refresh() {
			counter++
		}
	}()

	writeFiles(t)

	// help check this
	cancel()
	err = fcn.Stop()
	if err != nil {
		t.Fatal(err)
	}

	if got, want := counter, 3; got != want {
		t.Errorf("counter got %d want %d", got, want)
	}
}

func TestFileChangeNotifierRun(t *testing.T) {
	mkdirs(t)

	t.Cleanup(func() {
		_ = os.RemoveAll(dir1)
		_ = os.RemoveAll(dir2)
	})

	ctx, cancel := context.WithCancel(context.Background())
	fcn, err := NewFileChangeNotifier(
		ctx,
		[]DirFilesDescriptor{
			DirFilesDescriptor{dir1, []string{".html"}},
			DirFilesDescriptor{dir2, []string{"txt"}},
		},
	)
	if err != nil {
		t.Fatalf("error initialising fcn: %v", err)
	}

	counter := 0
	go func() {
		for {
			_ = fcn.Run() // blocking
			counter++
		}
	}()

	writeFiles(t)

	// help check this
	cancel()
	err = fcn.Stop()
	if err != nil {
		t.Fatal(err)
	}

	if got, want := counter, 3; got != want {
		t.Errorf("counter got %d want %d", got, want)
	}
}
