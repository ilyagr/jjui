package revisions

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/cellbuf"
	"github.com/idursun/jjui/internal/jj"
	"github.com/idursun/jjui/internal/parser"
	"github.com/idursun/jjui/internal/screen"
	"github.com/idursun/jjui/internal/ui/common"
	"github.com/idursun/jjui/internal/ui/intents"
	"github.com/idursun/jjui/internal/ui/layout"
	"github.com/idursun/jjui/internal/ui/operations"
	"github.com/idursun/jjui/internal/ui/render"
	"github.com/idursun/jjui/test"
	"github.com/stretchr/testify/assert"
)

var searchableRows = []parser.Row{
	{
		Commit: &jj.Commit{ChangeId: "first", CommitId: "111"},
		Lines: []*parser.GraphRowLine{
			{
				Gutter:   parser.GraphGutter{Segments: []*screen.Segment{{Text: "|"}}},
				Segments: []*screen.Segment{{Text: "first match"}},
				Flags:    parser.Revision,
			},
		},
	},
	{
		Commit: &jj.Commit{ChangeId: "second", CommitId: "222"},
		Lines: []*parser.GraphRowLine{
			{
				Gutter:   parser.GraphGutter{Segments: []*screen.Segment{{Text: "|"}}},
				Segments: []*screen.Segment{{Text: "second match"}},
				Flags:    parser.Revision,
			},
		},
	},
	{
		Commit: &jj.Commit{ChangeId: "third", CommitId: "333"},
		Lines: []*parser.GraphRowLine{
			{
				Gutter:   parser.GraphGutter{Segments: []*screen.Segment{{Text: "|"}}},
				Segments: []*screen.Segment{{Text: "third match"}},
				Flags:    parser.Revision,
			},
		},
	},
}

// mockNonFocusableOperation is a mock operation that is never focused, editing, or overlay
type mockNonFocusableOperation struct{}

func (m *mockNonFocusableOperation) Render(commit *jj.Commit, renderPosition operations.RenderPosition) string {
	return ""
}

func (m *mockNonFocusableOperation) RenderToDisplayContext(_ *render.DisplayContext, _ *jj.Commit, _ operations.RenderPosition, _ cellbuf.Rectangle, _ cellbuf.Position) int {
	return 0
}

func (m *mockNonFocusableOperation) DesiredHeight(_ *jj.Commit, _ operations.RenderPosition) int {
	return 0
}

func (m *mockNonFocusableOperation) Name() string {
	return "mock"
}

func (m *mockNonFocusableOperation) Init() tea.Cmd {
	return nil
}

func (m *mockNonFocusableOperation) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (m *mockNonFocusableOperation) ViewRect(_ *render.DisplayContext, _ layout.Box) {}

func (m *mockNonFocusableOperation) IsFocused() bool {
	return false
}

func (m *mockNonFocusableOperation) IsEditing() bool {
	return false
}

func (m *mockNonFocusableOperation) IsOverlay() bool {
	return false
}

// TestQuickSearch_EnterKeyClearsSearch tests that Enter clears the search
func TestQuickSearch_EnterKeyClearsSearch(t *testing.T) {
	model := &Model{
		quickSearch: "test",
		op:          &mockNonFocusableOperation{},
		rows:        []parser.Row{{Commit: &jj.Commit{ChangeId: "test123"}}},
	}

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	cmd := model.internalUpdate(msg)

	assert.Equal(t, "", model.quickSearch, "KeyEnter should clear quicksearch")
	assert.Nil(t, cmd)
}

// TestQuickSearch_EscapeKeyClearsSearch tests that Escape clears the search
func TestQuickSearch_EscapeKeyClearsSearch(t *testing.T) {
	model := &Model{
		quickSearch: "test",
		op:          &mockNonFocusableOperation{},
		rows:        []parser.Row{{Commit: &jj.Commit{ChangeId: "test123"}}},
	}

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	cmd := model.internalUpdate(msg)

	assert.Equal(t, "", model.quickSearch, "KeyEsc should clear quicksearch")
	assert.Nil(t, cmd)
}

// TestQuickSearch_EnterWithEmptySearch tests that Enter with empty search does nothing special
func TestQuickSearch_EnterWithEmptySearch(t *testing.T) {
	model := &Model{
		quickSearch: "",
		rows:        []parser.Row{{Commit: &jj.Commit{ChangeId: "test123"}}},
		op:          &mockNonFocusableOperation{},
	}

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_ = model.internalUpdate(msg)

	// When quickSearch is empty, Enter should not be handled by the quicksearch logic
	// The cmd might be nil or something else depending on other handlers
	assert.Equal(t, "", model.quickSearch)
}

func TestQuickSearch_UpdatesSelection(t *testing.T) {
	ctx := test.NewTestContext(test.NewTestCommandRunner(t))
	model := New(ctx)
	model.updateGraphRows(searchableRows, "first")

	selectionChanged := func(cmd tea.Cmd) bool {
		var changed bool
		test.SimulateModel(model, cmd, func(msg tea.Msg) {
			if _, ok := msg.(common.SelectionChangedMsg); ok {
				changed = true
			}
		})
		return changed
	}

	t.Run("QuickSearchMsg", func(t *testing.T) {
		assert.True(t, selectionChanged(model.Update(common.QuickSearchMsg("second"))))
	})

	t.Run("QuickSearchCycle", func(t *testing.T) {
		model.quickSearch = "match"
		assert.True(t, selectionChanged(model.Update(intents.QuickSearchCycle{})))
	})
}
