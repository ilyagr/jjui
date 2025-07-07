package status

import (
	"testing"

	"github.com/idursun/jjui/internal/ui/common"
	"github.com/idursun/jjui/internal/ui/context"
)

func TestRefreshBarAdvancesOnRefreshEvent(t *testing.T) {
	m := New(&context.MainContext{})
	mPtr := &m

	// Manually render the bar as in View()
	width := mPtr.refreshBar.Width
	pos := mPtr.refreshCount % width
	barRunes := make([]rune, width)
	for i := 0; i < width; i++ {
		if i == pos {
			barRunes[i] = '█'
		} else {
			barRunes[i] = ' '
		}
	}
	initial := string(barRunes)

	// Simulate a refresh event
	mPtr, _ = mPtr.Update(common.RefreshMsg{})

	// Render again
	width = mPtr.refreshBar.Width
	pos = mPtr.refreshCount % width
	barRunes = make([]rune, width)
	for i := 0; i < width; i++ {
		if i == pos {
			barRunes[i] = '█'
		} else {
			barRunes[i] = ' '
		}
	}
	after := string(barRunes)

	if initial == after {
		t.Errorf("Progress bar did not advance after refresh event: got %q, want different frame", after)
	}

	// Simulate another refresh event and check again
	initial = after
	mPtr, _ = mPtr.Update(common.RefreshMsg{})
	width = mPtr.refreshBar.Width
	pos = mPtr.refreshCount % width
	barRunes = make([]rune, width)
	for i := 0; i < width; i++ {
		if i == pos {
			barRunes[i] = '█'
		} else {
			barRunes[i] = ' '
		}
	}
	after = string(barRunes)
	if initial == after {
		t.Errorf("Progress bar did not advance after second refresh event: got %q, want different frame", after)
	}
}
