package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func writeFiles(t *testing.T, dir1, dir2 string, flushDuration time.Duration) {
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

func TestFileChangeCanonical(t *testing.T) {
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

	// override default flushDuration for testing
	fcn.flushDuration = 2 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	watchErr := make(chan error, 1)
	go func() {
		watchErr <- fcn.Watch(ctx)
	}()

	writeFiles(t, dir1, dir2, fcn.flushDuration)

	timeOut := time.After(2 * fcn.flushDuration)

	testOK := false
	select {
	case <-timeOut:
		t.Fatal("test timed out")
	case err := <-watchErr:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-fcn.Update():
		testOK = true
		cancel()
		break
	}

	if !testOK {
		t.Fatal("test failed")
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

	// override default flushDuration for testing
	fcn.flushDuration = 2 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	// 1. run Watch in it's own goroutine to check for watching errors.
	watchErr := make(chan error, 1)
	wg.Go(func() {
		watchErr <- fcn.Watch(ctx)
	})

	// 2. wait for Update events
	counter := 0
	wg.Go(func() {
		for range fcn.Update() {
			counter++
		}
	})

	// Write files then cancel the watcher.
	writeFiles(t, dir1, dir2, fcn.flushDuration)

	// Wait for last file to flush
	time.Sleep(2 * fcn.flushDuration)

	// Cancel Watch and Update (which share the context).
	cancel()

	// Wait for goroutines to finish
	wg.Wait()

	// check there are no watching errors
	err = <-watchErr
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected watch error: %v", err)
	}

	if got, want := counter, 3; got != want {
		t.Errorf("counter got %d want %d", got, want)
	}
}
