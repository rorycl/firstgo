package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeFiles(t *testing.T, dir1, dir2 string) {
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
		time.Sleep(flushDuration) // accommodate flush interval
	}
}

func TestFileChangeNotifier(t *testing.T) {

	dir1, dir2 := t.TempDir(), t.TempDir()

	fcn, err := NewFileChangeNotifier(
		[]DirFilesDescriptor{
			DirFilesDescriptor{dir1, []string{".html"}},
			DirFilesDescriptor{dir2, []string{"txt"}},
		},
	)
	if err != nil {
		t.Fatalf("error initialising fcn: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	err = fcn.Run(ctx)

	counter := 0
	go func() {
		for _ = range fcn.Refresh() {
			counter++
		}
	}()

	// write files then cancel the watcher
	flushDuration = 5 * time.Millisecond
	writeFiles(t, dir1, dir2)
	time.Sleep(flushDuration)
	cancel()

	if got, want := counter, 3; got != want {
		t.Errorf("counter got %d want %d", got, want)
	}
}
