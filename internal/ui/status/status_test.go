package status

import (
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/idursun/jjui/internal/ui/common"
	"github.com/idursun/jjui/internal/ui/context"
)

type dummyContext struct{}

func TestRefreshSpinnerAdvancesOnRefreshEvent(t *testing.T) {
	// Use a simple spinner with known frames for deterministic test
	testSpinner := spinner.New()
	testSpinner.Spinner = spinner.Line
	testSpinner.Style = testSpinner.Style // no-op, just to avoid nil

	// Create status model with dummy context
	m := New(&context.MainContext{})
	m.refreshSpinner = testSpinner

	// Use pointer receiver for Model
	mPtr := &m

	// Capture initial spinner frame
	initial := mPtr.refreshSpinner.View()

	// Simulate a refresh event
	// Bubble Tea Update returns (*Model, tea.Cmd)
	mPtr, _ = mPtr.Update(common.RefreshMsg{})

	// Spinner should have advanced by one frame
	after := mPtr.refreshSpinner.View()
	if initial == after {
		t.Errorf("Spinner did not advance after refresh event: got %q, want different frame", after)
	}
}
