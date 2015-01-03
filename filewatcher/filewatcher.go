/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package filewatcher

import (
	"fmt"
	"github.com/golang/glog"
	"gopkg.in/fsnotify.v1"
	"os"
	"path/filepath"
)

//####################//
//### Event Struct ###//
//####################//

type Event struct {
	Path string

	event *fsnotify.Event
}

func (e *Event) IsCreate() bool {
	return e.event.Op&fsnotify.Create == fsnotify.Create
}

func (e *Event) IsWrite() bool {
	return e.event.Op&fsnotify.Write == fsnotify.Write
}

func (e *Event) IsRename() bool {
	return e.event.Op&fsnotify.Rename == fsnotify.Rename
}

func (e *Event) IsRemove() bool {
	return e.event.Op&fsnotify.Remove == fsnotify.Remove
}

func (e *Event) IsChmod() bool {
	return e.event.Op&fsnotify.Chmod == fsnotify.Chmod
}

//##########################//
//### FileWatcher Struct ###//
//##########################//

type OnEventFunc func(event *Event)

type FileWatcher struct {
	watcher      *fsnotify.Watcher
	triggerClose chan struct{}
	onEvent      OnEventFunc
}

// Add starts watching the named file or directory
func (f *FileWatcher) Add(path string) error {
	// Clean the file path
	path = filepath.Clean(path)

	// Watch all sub directories
	err := filepath.Walk(path, func(path string, fi os.FileInfo, err error) error {
		if fi == nil {
			return fmt.Errorf("filepath walk: file info object is nil!")
		}

		if fi.IsDir() {
			err = f.watcher.Add(path)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	// Add the root path
	return f.watcher.Add(path)
}

// Remove stops watching the the named file or directory
func (f *FileWatcher) Remove(path string) error {
	// Clean the file path
	path = filepath.Clean(path)

	// TODO: Also remove recursive added directory paths added with Add!

	return f.watcher.Remove(path)
}

// Close removes all watchers and closes the events channel
func (f *FileWatcher) Close() {
	// Close the channel. This will exit the goroutine.
	close(f.triggerClose)
}

// OnEvent sets the function which is triggered on any filesystem event
func (f *FileWatcher) OnEvent(fn OnEventFunc) {
	f.onEvent = fn
}

//##############//
//### Public ###//
//##############//

func New() (*FileWatcher, error) {
	// Create a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	f := &FileWatcher{
		watcher:      watcher,
		triggerClose: make(chan struct{}),
	}

	// Start the goroutine to watch the files
	go func() {
		defer f.watcher.Close()

		for {
			select {
			case event := <-watcher.Events:
				// Skip if the event function is not defined
				if f.onEvent == nil {
					continue
				}

				// Create a new event object
				e := &Event{
					Path:  event.Name,
					event: &event,
				}

				// Call the callback in a safe way
				func() {
					// Recover panics and log the error
					defer func() {
						if e := recover(); e != nil {
							glog.Errorf("filewatcher callback panic: %v", e)
						}
					}()

					f.onEvent(e)
				}()

				// If a new folder was created, then add it to monitor recursive actions
				if e.IsCreate() {
					go func() {
						if stat, err := os.Stat(e.Path); err == nil && stat.IsDir() {
							err = f.Add(e.Path)
							if err != nil {
								glog.Errorf("filewatcher: failed to add recursive directory: %v", err)
							}
						}
					}()
				}

				// We don't have to remove deleted directories from the watcher,
				// because they are removed automatically...
			case err := <-watcher.Errors:
				glog.Errorf("a file watcher error occurred: %s", err.Error())
			case <-f.triggerClose:
				// Just exit the loop
				return
			}
		}
	}()

	return f, nil
}
