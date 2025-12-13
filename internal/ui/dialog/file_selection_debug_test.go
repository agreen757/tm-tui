package dialog

import (
	"path/filepath"
	"testing"
)

func TestReadDirectoryEntries_Debug(t *testing.T) {
	path := filepath.Join("/Users/adriangreen/Work/taskmaster-crush-fork", ".taskmaster", "docs")
	entries, err := readDirectoryEntries(path, map[string]struct{}{".md": {}, ".txt": {}})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("entries: %+v", entries)
}
