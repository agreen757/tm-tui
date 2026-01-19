package dialog

import (
    "os"
    "testing"
    "time"
)

func TestLoadPRDFile(t *testing.T) {
    // Create log viewer
    style := DefaultDialogStyle()
    viewer := NewLogViewerPanel(80, 24, style)
    
    // Get the file path relative to test
    prdPath := "../../../.taskmaster/docs/CLOUD_EXECUTION_PRD.md"
    
    // Check if file exists
    if _, err := os.Stat(prdPath); os.IsNotExist(err) {
        t.Skipf("PRD file not found at %s, skipping test", prdPath)
    }
    
    done := make(chan error, 1)
    
    go func() {
        err := viewer.LoadFileContent(prdPath)
        done <- err
    }()
    
    select {
    case err := <-done:
        if err != nil {
            t.Fatalf("LoadFileContent failed: %v", err)
        }
        t.Log("File loaded successfully!")
        t.Logf("Line limited: %v", viewer.IsLineLimited())
        t.Logf("File size warning: %v", viewer.HasFileSizeWarning())
    case <-time.After(10 * time.Second):
        t.Fatal("LoadFileContent timed out - likely infinite loop")
    }
    
    // Test rendering
    t.Log("Testing render...")
    viewer.SetFocused(true)
    view := viewer.View()
    t.Logf("Render output: %d chars", len(view))
    
    // Test with markdown enabled
    t.Log("Testing markdown render...")
    viewer.ToggleMarkdown()
    view = viewer.View()
    t.Logf("Markdown render output: %d chars", len(view))
}
