package taskmaster

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// UndoActionType enumerates supported undo operations.
type UndoActionType string

const (
	// UndoActionDelete is used when tasks were deleted.
	UndoActionDelete UndoActionType = "delete_tasks"
)

// ErrUndoExpired indicates the undo action has expired.
var ErrUndoExpired = errors.New("undo action expired")

// ErrUndoNotFound indicates no undo action is available for the supplied ID.
var ErrUndoNotFound = errors.New("undo action not found")

// UndoToken is a lightweight handle surfaced to callers so they can show UX cues.
type UndoToken struct {
	ID        string
	Type      UndoActionType
	Summary   string
	ExpiresAt time.Time
	Duration  time.Duration
}

// UndoAction holds the data required to rollback an operation.
type UndoAction struct {
	ID        string
	Type      UndoActionType
	Summary   string
	ExpiresAt time.Time
	Payload   []byte
	Metadata  map[string]string
}

// Token creates a user-facing token for the undo action.
func (a *UndoAction) Token() *UndoToken {
	if a == nil {
		return nil
	}
	return &UndoToken{
		ID:        a.ID,
		Type:      a.Type,
		Summary:   a.Summary,
		ExpiresAt: a.ExpiresAt,
		Duration:  time.Until(a.ExpiresAt),
	}
}

// UndoManager stores the latest undoable action with simple expiration support.
type UndoManager struct {
	mu     sync.Mutex
	action *UndoAction
}

// NewUndoManager creates a fresh undo manager.
func NewUndoManager() *UndoManager {
	return &UndoManager{}
}

// Push replaces the current undo action with the provided one.
func (m *UndoManager) Push(action *UndoAction) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.action = action
}

// Peek returns the current undo action without consuming it.
func (m *UndoManager) Peek() *UndoAction {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.action == nil {
		return nil
	}
	if time.Now().After(m.action.ExpiresAt) {
		m.action = nil
		return nil
	}
	return m.action
}

// Consume removes and returns the current undo action if the ID matches and it
// has not expired.
func (m *UndoManager) Consume(id string) (*UndoAction, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.action == nil {
		return nil, ErrUndoNotFound
	}
	if m.action.ID != id {
		return nil, fmt.Errorf("%w: %s", ErrUndoNotFound, id)
	}
	if time.Now().After(m.action.ExpiresAt) {
		m.action = nil
		return nil, ErrUndoExpired
	}
	action := m.action
	m.action = nil
	return action, nil
}

// newDeleteUndoAction creates an undo action for delete workflows.
func newDeleteUndoAction(payload []byte, deletedCount int, ttl time.Duration) *UndoAction {
	if ttl <= 0 {
		ttl = 20 * time.Second
	}
	return &UndoAction{
		ID:        fmt.Sprintf("undo-%d", time.Now().UnixNano()),
		Type:      UndoActionDelete,
		Summary:   fmt.Sprintf("Deleted %d task(s)", deletedCount),
		ExpiresAt: time.Now().Add(ttl),
		Payload:   payload,
		Metadata: map[string]string{
			"deletedCount": fmt.Sprintf("%d", deletedCount),
		},
	}
}
