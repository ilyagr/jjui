package status

import (
	"testing"

	"github.com/idursun/jjui/internal/ui/common"
	"github.com/idursun/jjui/internal/ui/context"
)

func TestRefreshBarAdvancesOnRefreshEvent(t *testing.T) {
	m := New(&context.MainContext{})
	mPtr := &m

	// Get initial spinner character
	initial := mPtr.spinnerChars[mPtr.spinnerIdx]

	// Simulate a refresh completion event
	mPtr, _ = mPtr.Update(common.UpdateRevisionsSuccessMsg{})

	// Get spinner character after refresh completion
	after := mPtr.spinnerChars[mPtr.spinnerIdx]
	if initial == after {
		t.Errorf("Spinner did not advance after refresh completion event: got %q, want different frame", string(after))
	}

	// Simulate another refresh completion event and check again
	initial = after
	mPtr, _ = mPtr.Update(common.UpdateRevisionsSuccessMsg{})
	after = mPtr.spinnerChars[mPtr.spinnerIdx]
	if initial == after {
		t.Errorf("Spinner did not advance after second refresh completion event: got %q, want different frame", string(after))
	}
}
