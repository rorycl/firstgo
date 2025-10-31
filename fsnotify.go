package main

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"golang.org/x/sync/errgroup"
)

// flushDuration sets the time given to wait for multiple editor writes
const flushDuration time.Duration = 50 * time.Millisecond

// DirFilesDescriptor is a combination of a directory and files with the
// specified suffixes to watch under it.
type DirFilesDescriptor struct {
	Dir          string
	FileSuffixes []string
}

// FileChangeNotifier is a type holding one or more FileChangeDescriptor
// watchers.
type FileChangeNotifier struct {
	dirFiles         []DirFilesDescriptor
	dirDescriptorMap map[string][]string
	watcher          *fsnotify.Watcher
	refresh          chan bool
	err              error
}

// Refresh reports need for a refresh providing an iterator of bool,
// error to catch both file change events and possible errors that may
// occur during the file watching process. No error is deemed fatal and
// it is up to the the consumer to catch errors and exit the loop
// appropriately.
func (fcn *FileChangeNotifier) Refresh() iter.Seq2[bool, error] {
	return func(yield func(bool, error) bool) {
		for r := range fcn.refresh {
			if !yield(r, fcn.err) {
				return
			}
		}
	}
}

// NewFileChangeNotifier registers and starts a FileChangeNotifier,
// watching the specified directories for write events for files with
// the specified suffixes. Consumers should iterate over [Refresh] to
// receive events or errors.
//
// Note that suffixes provided without the leading "dot" ('.') have this
// prepended to the provided suffix.
//
// Refer to
// https://github.com/fsnotify/fsnotify/blob/v1.8.0/cmd/fsnotify/file.go
//
// An example:
//
//	func main() {
//		watcher, err := NewFileChangeNotifier(
//			context.TODO(),
//			[]DirFilesDescriptor{
//				DirFilesDescriptor{"/tmp/a", []string{".html", "css"}},
//				DirFilesDescriptor{"/tmp/b", []string{"txt"}},
//			},
//		)
//		if err != nil {
//			fmt.Println("error at base:", err)
//			os.Exit(1)
//		}
//		for _, err = range watcher.Refresh() {
//			fmt.Println(err, "got!")
//		}
//	}
func NewFileChangeNotifier(ctx context.Context, descriptors []DirFilesDescriptor) (*FileChangeNotifier, error) {

	if len(descriptors) < 1 {
		return nil, fmt.Errorf("need at least one dir/filematch descriptor")
	}

	fcn := FileChangeNotifier{
		dirFiles:         descriptors,
		dirDescriptorMap: map[string][]string{},
		refresh:          make(chan bool),
	}

	var err error
	fcn.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("fsnotify new watcher error: %w", err)
	}

	for _, desc := range fcn.dirFiles {
		dir := filepath.Clean(desc.Dir)
		check, err := os.Stat(dir)
		if err != nil {
			return nil, fmt.Errorf("dir %q not found: %w", dir, err)
		}
		if !check.IsDir() {
			return nil, fmt.Errorf("%q is not a directory", dir)
		}
		if _, found := fcn.dirDescriptorMap[dir]; found {
			return nil, fmt.Errorf("%q already registered", dir)
		}
		err = fcn.watcher.Add(dir)
		if err != nil {
			return nil, fmt.Errorf("fsnotify add error for dir %q: %w", dir, err)
		}

		// add the suffixes, prepending "." if necessary.
		fcn.dirDescriptorMap[dir] = []string{}
		for _, ix := range desc.FileSuffixes {
			if len(ix) > 0 && ix[0] != byte('.') {
				ix = string('.') + ix
			}
			fcn.dirDescriptorMap[dir] = append(fcn.dirDescriptorMap[dir], ix)
		}
	}

	// internal eventChan (used for buffering)
	eventChan := make(chan bool)

	g, ctx := errgroup.WithContext(ctx)

	// This goroutine watches for *fsnotify.Watcher events.
	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case err, ok := <-fcn.watcher.Errors:
				if !ok {
					return errors.New("unexpected close from watcher.Errors")
				}
				return fmt.Errorf("unexpected notify error: %w", err)

			// An event has been received.
			case e, ok := <-fcn.watcher.Events:
				if !ok {
					return errors.New("unexpected close from watcher.Events")
				}
				// skip events that aren't writes
				if !e.Has(fsnotify.Write) {
					continue
				}
				dir := filepath.Dir(e.Name)
				basename := filepath.Base(e.Name)
				// fmt.Printf("event for %s\n    string: %s\n", e.Name, e.String())

				// ignore dot files
				if len(basename) > 0 && basename[0] == '.' {
					continue
				}

				// check the suffixes for this directory
				suffixes, ok := fcn.dirDescriptorMap[dir]
				if !ok {
					return fmt.Errorf("could not find matcher for dir %q", dir)
				}
				for _, ix := range suffixes {
					if strings.HasSuffix(strings.ToLower(basename), strings.ToLower(ix)) {
						eventChan <- true
					}
				}
			}
		}
	})

	// Simple buffer of double writes by editors like vim. This
	// goroutine will exit if the context is Done or eventChan is
	// closed.
	g.Go(func() error {
		flush := false
		timer := time.NewTicker(flushDuration)
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
				// Stack writes in the same flushDuration, giving time for
				// the writes to complete.
			case _, ok := <-eventChan:
				if !ok {
					return nil
				}
				flush = true
				timer.Reset(flushDuration)
			case <-timer.C:
				if flush {
					fcn.refresh <- true
					flush = false
				}
			}
		}
	})

	go func() {
		fcn.err = g.Wait()
		close(eventChan)
		_ = fcn.watcher.Close()
	}()

	return &fcn, nil
}
