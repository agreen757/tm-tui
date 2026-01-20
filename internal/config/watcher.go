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

	// Create reusable timer
	debounceTimer := time.NewTimer(debounceInterval)
	// Drain initial timer if it fires
	if !debounceTimer.Stop() {
		<-debounceTimer.C
	}
	
	var mu sync.Mutex
	var pendingEvent bool
	var timerActive bool

	for {
		select {
		case <-w.ctx.Done():
			mu.Lock()
			debounceTimer.Stop()
			mu.Unlock()
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

				mu.Lock()
				pendingEvent = true

				// Reset debounce timer
				if timerActive {
					if !debounceTimer.Stop() {
						select {
						case <-debounceTimer.C:
						default:
						}
					}
				}
				debounceTimer.Reset(debounceInterval)
				timerActive = true
				mu.Unlock()
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
			
		case <-debounceTimer.C:
			mu.Lock()
			if pendingEvent {
				select {
				case w.events <- struct{}{}:
				default:
					// Channel full, event already pending
				}
				pendingEvent = false
			}
			timerActive = false
			mu.Unlock()
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
		
		// Use explicit timer with proper cleanup
		timer := time.NewTimer(interval)
		// Drain initial timer if fired
		if !timer.Stop() {
			<-timer.C
		}
		
		var mu sync.Mutex
		var pendingEvent bool
		var timerRunning bool

		for {
			select {
			case _, ok := <-input:
				if !ok {
					// Input closed, set pending flag and stop timer
					mu.Lock()
					timerRunning = false
					pendingEvent = false
					timer.Stop()
					mu.Unlock()
					return
				}
				
				mu.Lock()
				pendingEvent = true

				// Reset timer if not running
				if !timerRunning {
					timer.Reset(interval)
					timerRunning = true
				} else {
					// Stop and restart to reset the interval
					if !timer.Stop() {
						select {
						case <-timer.C:
						default:
						}
					}
					timer.Reset(interval)
				}
				mu.Unlock()

			case <-timer.C:
				mu.Lock()
				if pendingEvent {
					select {
					case output <- struct{}{}:
					default:
						// Channel buffer full, skip this one
					}
					pendingEvent = false
				}
				timerRunning = false
				mu.Unlock()
			}
		}
	}()

	return output
}
