package revisions

import (
	"io"
	"strconv"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/idursun/jjui/internal/config"
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
	// Use the actual keymap type expected by Model
	testKeymap := getTestKeyMap()

	model := &Model{
		rows:              rows,
		cursor:            0,
		selectedRevisions: make(map[string]bool),
		viewRange:         &viewRange{start: 0, end: 0, lastRowIndex: -1},
		keymap:            testKeymap,
	}

	// Wrap Model to implement tea.Model interface
	wrapped := &modelTeaWrapper{Model: model}

	catwalk.RunModelFromString(t, `
run
----
-- view:
a
b
c

run
key down
----
-- gostruct:
cursor: 1

run
msg refresh
----
-- gostruct:
cursor: 1

run
key down
----
-- gostruct:
cursor: 2

run
msg refresh
----
-- gostruct:
cursor: 2
`, wrapped,
		catwalk.WithObserver("gostruct", func(out io.Writer, m tea.Model) error {
			if mm, ok := m.(*modelTeaWrapper); ok {
				_, _ = out.Write([]byte(
					"cursor: " + itoa(mm.Model.cursor) + "\n",
				))
			}
			return nil
		}),
		catwalk.WithUpdater(func(m tea.Model, cmd string, args ...string) (bool, tea.Model, tea.Cmd, error) {
			if cmd == "msg" && len(args) == 1 && args[0] == "refresh" {
				if mm, ok := m.(*modelTeaWrapper); ok {
					return true, mm, func() tea.Msg { return struct{ KeepSelections bool }{} }, nil
				}
			}
			return false, nil, nil, nil
		}),
	)
}

// modelTeaWrapper wraps *Model to implement tea.Model interface for catwalk.
type modelTeaWrapper struct {
	Model *Model
}

func (w *modelTeaWrapper) Init() tea.Cmd {
	return nil
}

func (w *modelTeaWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := w.Model.Update(msg)
	w.Model = m
	return w, cmd
}

func (w *modelTeaWrapper) View() string {
	return w.Model.View()
}

// itoa is a minimal int to string for small ints (0-9)
func itoa(i int) string {
	return strconv.Itoa(i)
}

// getTestKeyMap returns a minimal keymap for testing.
func getTestKeyMap() config.KeyMappings[key.Binding] {
	// Use the real config.Convert to get a KeyMappings[key.Binding]
	return config.Convert(config.DefaultKeyMappings)
}
