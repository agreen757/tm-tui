package ui

import (
	"testing"

	"github.com/adriangreen/tm-tui/internal/ui/dialog"
)

func newTestModel() Model {
	keyMap := NewKeyMap(nil)
	dm := dialog.InitializeDialogManager(120, 60, dialog.DefaultDialogStyle())
	return Model{
		appState: NewAppState(dm, &keyMap),
		keyMap:   keyMap,
		commands: defaultCommandSpecs(),
	}
}

func TestCommandDispatchOpensDialogs(t *testing.T) {
	model := newTestModel()
	cases := []struct {
		command CommandID
		title   string
	}{
		{CommandParsePRD, "Select PRD File"},
		{CommandAnalyzeComplexity, "Analyze Task Complexity"},
		{CommandManageTags, "Add Tag Context"},
		{CommandUseTag, "Use Tag Context"},
	}

	for _, tc := range cases {
		t.Run(string(tc.command), func(t *testing.T) {
			model.appState.ClearDialogs()
			model.dispatchCommand(tc.command)
			dialog := model.appState.ActiveDialog()
			if dialog == nil {
				t.Fatalf("expected a dialog for %s but none was active", tc.command)
			}
			if dialog.Title() != tc.title {
				t.Fatalf("expected dialog title %q, got %q", tc.title, dialog.Title())
			}
		})
	}
}

func TestExpandTaskCommandShowsErrorWithoutSelection(t *testing.T) {
	model := newTestModel()
	model.appState.ClearDialogs()

	model.dispatchCommand(CommandExpandTask)

	active := model.appState.ActiveDialog()
	if active == nil {
		t.Fatal("expected error dialog when expanding without selection")
	}

	errDialog, ok := active.(*dialog.ErrorDialogModel)
	if !ok {
		t.Fatalf("expected ErrorDialogModel, got %T", active)
	}

	if errDialog.Style == nil || errDialog.Style.BorderColor != errDialog.Style.ErrorColor {
		t.Fatalf("expected error dialog style to use error color border")
	}
}
