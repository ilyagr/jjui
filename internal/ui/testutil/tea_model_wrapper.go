package testutil

import tea "github.com/charmbracelet/bubbletea"

// TeaModelWrapper wraps any tea.Model to allow pointer mutation in tests.
// This enables tests to use a generic adapter for any model implementing tea.Model.
type TeaModelWrapper struct {
	Model tea.Model
}

func (w *TeaModelWrapper) Init() tea.Cmd {
	return w.Model.Init()
}

func (w *TeaModelWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := w.Model.Update(msg)
	w.Model = m
	return w, cmd
}

func (w *TeaModelWrapper) View() string {
	return w.Model.View()
}
