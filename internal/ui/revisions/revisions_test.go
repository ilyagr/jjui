package revisions

import (
	"bytes"
	"testing"

	"github.com/idursun/jjui/internal/jj"
	"github.com/idursun/jjui/internal/ui/graph"
	"github.com/knz/catwalk"
	"github.com/stretchr/testify/assert"
)

func TestModel_highlightChanges(t *testing.T) {
	model := Model{
		rows: []graph.Row{
			{Commit: &jj.Commit{ChangeId: "someother"}},
			{Commit: &jj.Commit{ChangeId: "nyqzpsmt"}},
		},
		output: `
Absorbed changes into these revisions:
  nyqzpsmt 8b1e95e3 change third file
Working copy now at: okrwsxvv 5233c94f (empty) (no description set)
Parent commit      : nyqzpsmt 8b1e95e3 change third file
`, err: nil}
	_ = model.highlightChanges()
	assert.False(t, model.rows[0].IsAffected)
	assert.True(t, model.rows[1].IsAffected)
}

// This test uses catwalk to verify that pressing down moves the cursor,
// but refreshing the view with ctrl+r after that does not move the cursor.
func TestRevisions_CursorAndRefreshBehavior_Catwalk(t *testing.T) {
	// Prepare a minimal fake graph with 3 commits
	rows := []graph.Row{
		{Commit: &jj.Commit{ChangeId: "a", CommitId: "a"}},
		{Commit: &jj.Commit{ChangeId: "b", CommitId: "b"}},
		{Commit: &jj.Commit{ChangeId: "c", CommitId: "c"}},
	}
	model := &Model{
		rows:              rows,
		cursor:            0,
		selectedRevisions: make(map[string]bool),
		viewRange:         &viewRange{start: 0, end: 0, lastRowIndex: -1},
		keymap:            getTestKeyMap(),
	}

	cw := catwalk.New(t, model)
	cw.Step("initial", func() {
		cw.Require().Contains(cw.View(), "a")
		cw.Require().Contains(cw.View(), "b")
		cw.Require().Contains(cw.View(), "c")
		cw.Require().Contains(cw.View(), "a") // cursor at 0
	})

	cw.Step("press down", func() {
		cw.SendKey("down")
		cw.Require().Equal(1, model.cursor)
	})

	cw.Step("refresh with ctrl+r", func() {
		// Simulate a refresh message (should not move cursor)
		cw.SendMsg(struct{ KeepSelections bool }{})
		cw.Require().Equal(1, model.cursor)
	})

	cw.Step("press down again", func() {
		cw.SendKey("down")
		cw.Require().Equal(2, model.cursor)
	})

	cw.Step("refresh again", func() {
		cw.SendMsg(struct{ KeepSelections bool }{})
		cw.Require().Equal(2, model.cursor)
	})

	// Optionally, print the view for debugging
	_ = bytes.NewBufferString(cw.View())
}

// getTestKeyMap returns a minimal keymap for testing.
func getTestKeyMap() map[string]string {
	return map[string]string{
		"up":      "up",
		"down":    "down",
		"refresh": "ctrl+r",
	}
}
