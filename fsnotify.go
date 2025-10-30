package main

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"golang.org/x/sync/errgroup"
)

// flushDuration sets the time given to wait for multiple editor writes
const flushDuration time.Duration = 50 * time.Millisecond

// FileChangeDescriptor is a combination of a directory and file names
// to watch under it.
type DirFilesDescriptor struct {
	Dir         string
	FileMatcher *regexp.Regexp
}

// FileChangeNotifier is a type holding one or more FileChangeDescriptor
// watchers.
type FileChangeNotifier struct {
	dirFiles         []DirFilesDescriptor
	dirDescriptorMap map[string]*regexp.Regexp
	watcher          *fsnotify.Watcher
	refresh          chan bool
	Err              error
	sync.Mutex
}

// Refresh reports need for a refresh, enclosing a bool channel.
func (fcn *FileChangeNotifier) Refresh() iter.Seq[bool] {
	return func(yield func(bool) bool) {
		for r := range fcn.refresh {
			if !yield(r) {
				return
			}
		}
	}
}

// NewFileChangeNotifier registers and starts a FileChangeNotifier,
// watching the specified directories for events. This code refers
// closely to
// https://github.com/fsnotify/fsnotify/blob/v1.8.0/cmd/fsnotify/file.go
func NewFileChangeNotifier(ctx context.Context, descriptors []DirFilesDescriptor) (*FileChangeNotifier, error) {

	if len(descriptors) < 1 {
		return nil, fmt.Errorf("need at least one dir/filematch descriptor")
	}

	fcn := FileChangeNotifier{
		dirFiles:         descriptors,
		dirDescriptorMap: map[string]*regexp.Regexp{},
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
		fcn.dirDescriptorMap[dir] = desc.FileMatcher
	}

	// internal eventChan
	eventChan := make(chan bool)

	g := new(errgroup.Group)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case err, ok := <-fcn.watcher.Errors:
				if !ok {
					errRecorder(errors.New("unexpected close from watcher.Errors"))
					return
				}
				errRecorder(fmt.Errorf("unexpected notify error: %w", err))
				return

			case e, ok := <-fcn.watcher.Events:
				if !ok {
					errRecorder(errors.New("unexpected close from watcher.Events"))
					return
				}
				// skip events that aren't writes
				if !e.Has(fsnotify.Write) {
					continue
				}
				dir := filepath.Dir(e.Name)
				basename := filepath.Base(e.Name)
				fmt.Printf("event for %s\n    string: %s\n", e.Name, e.String())

				// Find the file regexp filename matcher for the event
				// directory, erroring if it can't be found. Emit an
				// event if there is a match.
				matcher, ok := fcn.dirDescriptorMap[dir]
				if !ok {
					errRecorder(fmt.Errorf("could not find matcher for dir %q", dir))
				}
				if matcher.MatchString(basename) {
					eventChan <- true
				} else {
					fmt.Printf("%q no match for %q\n", matcher.String(), basename)
				}
			}
		}
	}()

	// Simple buffer of double writes by editors like vim. This
	// goroutine will exit if the context is Done or eventChan is
	// closed.
	go func() {
		flush := false
		timer := time.NewTicker(flushDuration)
		for {
			select {
			case <-ctx.Done():
				return
				// Stack writes in the same flushDuration, giving time for
				// the writes to complete.
			case _, ok := <-eventChan:
				if !ok {
					return
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
	}()

	go func() {
		wg.Wait()
		close(eventChan)
		defer fcn.watcher.Close()
	}()

	return &fcn, nil

}

func main() {

	watcher, err := NewFileChangeNotifier(
		context.TODO(),
		[]DirFilesDescriptor{
			DirFilesDescriptor{"/tmp/zan", regexp.MustCompile("(?i)^[a-z0-9].+html$")},
			DirFilesDescriptor{"/tmp/zib", regexp.MustCompile("^[A-Z].+txt$")},
		},
	)
	if err != nil {
		fmt.Println("error at base", err)
		os.Exit(1)
	}
	for _ = range watcher.Refresh() {
		fmt.Println("got!")
	}
	if watcher.Err != nil {
		fmt.Println("unexpected err", err)
	}
}
