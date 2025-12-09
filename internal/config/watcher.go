package config

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher wraps fsnotify to watch files and emit debounced change notifications
type Watcher struct {
	watcher  *fsnotify.Watcher
	paths    []string
	events   chan struct{}
	errors   chan error
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
	watching bool
}

// NewWatcher creates a new file watcher for the specified paths
func NewWatcher(ctx context.Context, paths ...string) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	watcherCtx, cancel := context.WithCancel(ctx)

	w := &Watcher{
		watcher:  fsw,
		paths:    paths,
		events:   make(chan struct{}, 1),
		errors:   make(chan error, 1),
		ctx:      watcherCtx,
		cancel:   cancel,
		watching: false,
	}

	return w, nil
}

// Start begins watching the configured paths with debouncing
func (w *Watcher) Start(debounceInterval time.Duration) error {
	w.mu.Lock()
	if w.watching {
		w.mu.Unlock()
		return fmt.Errorf("watcher already started")
	}
	w.watching = true
	w.mu.Unlock()

	// Add paths to watcher
	for _, path := range w.paths {
		// For file paths, watch the parent directory
		dir := filepath.Dir(path)
		if err := w.watcher.Add(dir); err != nil {
			w.watching = false
			return fmt.Errorf("failed to watch %s: %w", dir, err)
		}
	}

	// Start goroutine to process events
	go w.processEvents(debounceInterval)

	return nil
}

// processEvents handles fsnotify events and applies debouncing
func (w *Watcher) processEvents(debounceInterval time.Duration) {
	defer close(w.events)
	defer close(w.errors)

	var debounceTimer *time.Timer
	var pendingEvent bool

	for {
		select {
		case <-w.ctx.Done():
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			return

		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			// Check if this event is for one of our watched files
			if !w.isWatchedFile(event.Name) {
				continue
			}

			// Only process Write and Create events (ignoring Chmod, etc.)
			if event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Create == fsnotify.Create ||
				event.Op&fsnotify.Rename == fsnotify.Rename {

				pendingEvent = true

				// Reset debounce timer
				if debounceTimer != nil {
					debounceTimer.Stop()
				}

				debounceTimer = time.AfterFunc(debounceInterval, func() {
					if pendingEvent {
						select {
						case w.events <- struct{}{}:
						default:
							// Channel full, event already pending
						}
						pendingEvent = false
					}
				})
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			select {
			case w.errors <- err:
			case <-w.ctx.Done():
				return
			}
		}
	}
}

// isWatchedFile checks if the given path is one of the watched files
func (w *Watcher) isWatchedFile(path string) bool {
	for _, watchedPath := range w.paths {
		if filepath.Base(path) == filepath.Base(watchedPath) {
			return true
		}
	}
	return false
}

// Events returns the channel for receiving debounced file change notifications
func (w *Watcher) Events() <-chan struct{} {
	return w.events
}

// Errors returns the channel for receiving watcher errors
func (w *Watcher) Errors() <-chan error {
	return w.errors
}

// Stop stops the watcher and cleans up resources
func (w *Watcher) Stop() error {
	w.mu.Lock()
	if !w.watching {
		w.mu.Unlock()
		return nil
	}
	w.watching = false
	w.mu.Unlock()

	w.cancel()
	return w.watcher.Close()
}

// Debounce creates a debounced channel from an input channel
// This is a standalone utility that can be used independently
func Debounce(interval time.Duration, input <-chan struct{}) <-chan struct{} {
	output := make(chan struct{})

	go func() {
		defer close(output)
		var timer *time.Timer

		for range input {
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(interval, func() {
				output <- struct{}{}
			})
		}
	}()

	return output
}
