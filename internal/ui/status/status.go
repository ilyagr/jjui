package status

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/idursun/jjui/internal/config"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/idursun/jjui/internal/ui/common"
	"github.com/idursun/jjui/internal/ui/context"
	"github.com/idursun/jjui/internal/ui/exec_process"
	"github.com/idursun/jjui/internal/ui/fuzzy_files"
	"github.com/idursun/jjui/internal/ui/fuzzy_search"
)

var accept = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "accept"))

type commandStatus int

const (
	none commandStatus = iota
	commandRunning
	commandCompleted
	commandFailed
)

type Model struct {
	context      *context.MainContext
	refreshCount int           // Number of refreshes that have occurred
	spinnerChars []rune        // Spinner characters for single-cell spinner
	spinnerIdx   int           // Current spinner index
	spinner      spinner.Model // Existing spinner for external commands (deprecated for refresh)
	input        textinput.Model
	keyMap       help.KeyMap
	command      string
	status  commandStatus
	running      bool
	width        int
	mode         string
	editing      bool
	history      map[string][]string
	fuzzy   fuzzy_search.Model
	styles       styles
}

type styles struct {
	shortcut lipgloss.Style
	dimmed   lipgloss.Style
	text     lipgloss.Style
	title    lipgloss.Style
	success  lipgloss.Style
	error    lipgloss.Style
}

func (m *Model) FuzzyView() string {
	if m.fuzzy == nil {
		return ""
	}
	return m.fuzzy.View()
}

func (m *Model) IsFocused() bool {
	return m.editing
}

const CommandClearDuration = 3 * time.Second

type clearMsg string

func (m *Model) Width() int {
	return m.width
}

func (m *Model) Height() int {
	return 1
}

func (m *Model) SetWidth(w int) {
	m.width = w
}

func (m *Model) SetHeight(int) {}
func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	km := config.Current.GetKeyMap()
	switch msg := msg.(type) {
	case clearMsg:
		if m.command == string(msg) {
			m.command = ""
			m.status = none
		}
		return m, nil
	case common.CommandRunningMsg:
		m.command = string(msg)
		m.status = commandRunning
		return m, m.spinner.Tick
	case common.CommandCompletedMsg:
		if msg.Err != nil {
			m.status = commandFailed
		} else {
			m.status = commandCompleted
		}
		commandToBeCleared := m.command
		return m, tea.Tick(CommandClearDuration, func(time.Time) tea.Msg {
			return clearMsg(commandToBeCleared)
		})
	case common.FileSearchMsg:
		m.editing = true
		m.mode = "rev file"
		m.input.Prompt = "> "
		m.loadEditingSuggestions()
		m.fuzzy = fuzzy_files.NewModel(msg)
		return m, tea.Batch(m.fuzzy.Init(), m.input.Focus())
	case common.UpdateRevisionsSuccessMsg:
		// Advance spinner by one tick when refresh is done
		m.refreshCount++
		m.spinnerIdx = m.refreshCount % len(m.spinnerChars)
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, km.Cancel) && m.editing:
			var cmd tea.Cmd
			if m.fuzzy != nil {
				_, cmd = m.fuzzy.Update(msg)
			}

			m.fuzzy = nil
			m.editing = false
			m.input.Reset()
			return m, cmd
		case key.Matches(msg, accept) && m.editing:
			editMode := m.mode
			input := m.input.Value()
			prompt := m.input.Prompt
			fuzzy := m.fuzzy
			m.saveEditingSuggestions()

			m.fuzzy = nil
			m.command = ""
			m.editing = false
			m.mode = ""
			m.input.Reset()

			switch {
			case strings.HasSuffix(editMode, "file"):
				_, cmd := fuzzy.Update(msg)
				return m, cmd
			case strings.HasPrefix(editMode, "exec"):
				return m, func() tea.Msg { return exec_process.ExecMsgFromLine(prompt, input) }
			}
			return m, func() tea.Msg { return common.QuickSearchMsg(input) }
		case key.Matches(msg, km.ExecJJ, km.ExecShell) && !m.editing:
			mode := common.ExecJJ
			if key.Matches(msg, km.ExecShell) {
				mode = common.ExecShell
			}
			m.editing = true
			m.mode = "exec " + mode.Mode
			m.input.Prompt = mode.Prompt
			m.loadEditingSuggestions()
			return m, m.input.Focus()
		case key.Matches(msg, km.QuickSearch) && !m.editing:
			m.editing = true
			m.mode = "search"
			m.input.Prompt = "> "
			m.loadEditingSuggestions()
			return m, m.input.Focus()
		default:
			if m.editing {
				var cmd tea.Cmd
				m.input, cmd = m.input.Update(msg)
				if m.fuzzy != nil {
					cmd = tea.Batch(cmd, fuzzy_search.Search(m.input.Value(), msg))
				}
				return m, cmd
			}
		}
		return m, nil
	default:
		var cmd tea.Cmd
		// No-op for progress bar unless we want to animate it (we don't)
		// Still update command spinner for legacy/external commands
		if m.status == commandRunning {
			m.spinner, _ = m.spinner.Update(msg)
		}
		if m.fuzzy != nil {
			m.fuzzy, cmd = fuzzy_search.Update(m.fuzzy, msg)
		}
		return m, cmd
	}
}

func (m *Model) saveEditingSuggestions() {
	if h, ok := m.history[m.mode]; ok {
		m.history[m.mode] = append(h, m.input.Value())
	} else {
		m.history[m.mode] = []string{m.input.Value()}
	}
}

func (m *Model) loadEditingSuggestions() {
	if h, ok := m.history[m.mode]; ok {
		m.input.ShowSuggestions = true
		m.input.SetSuggestions(h)
	} else {
		m.input.ShowSuggestions = false
		m.input.SetSuggestions([]string{})
	}
}

func (m *Model) View() string {
	// Single-cell spinner using spinnerChars
	spinnerChar := m.spinnerChars[m.spinnerIdx]
	refreshSpinnerMark := m.styles.title.Render(string(spinnerChar))

	commandStatusMark := m.styles.text.Render(" ")
	if m.status == commandRunning {
		commandStatusMark = m.styles.text.Render(m.spinner.View())
	} else if m.status == commandFailed {
		commandStatusMark = m.styles.error.Render("✗ ")
	} else if m.status == commandCompleted {
		commandStatusMark = m.styles.success.Render("✓ ")
	} else {
		commandStatusMark = m.helpView(m.keyMap)
	}
	ret := m.styles.text.Render(strings.ReplaceAll(m.command, "\n", "⏎"))
	if m.editing {
		commandStatusMark = ""
		ret = m.input.View()
	}
	mode := m.styles.title.Width(10).Render("", m.mode)
	// Place refresh spinner to the left of the mode indicator
	ret = lipgloss.JoinHorizontal(lipgloss.Left, refreshSpinnerMark, mode,  m.styles.text.Render(" "), commandStatusMark, ret)
	height := lipgloss.Height(ret)
	return lipgloss.Place(m.width, height, 0, 0, ret, lipgloss.WithWhitespaceBackground(m.styles.text.GetBackground()))
}

func (m *Model) SetHelp(keyMap help.KeyMap) {
	m.keyMap = keyMap
}

func (m *Model) SetMode(mode string) {
	if !m.editing {
		m.mode = mode
	}
}

func (m *Model) helpView(keyMap help.KeyMap) string {
	shortHelp := keyMap.ShortHelp()
	var entries []string
	for _, binding := range shortHelp {
		if !binding.Enabled() {
			continue
		}
		h := binding.Help()
		entries = append(entries, m.styles.shortcut.Render(h.Key)+m.styles.dimmed.PaddingLeft(1).Render(h.Desc))
	}
	return lipgloss.PlaceHorizontal(m.width, 0, strings.Join(entries, m.styles.dimmed.Render(" • ")), lipgloss.WithWhitespaceBackground(m.styles.text.GetBackground()))
}

func New(context *context.MainContext) Model {
	styles := styles{
		shortcut: common.DefaultPalette.Get("status shortcut"),
		dimmed:   common.DefaultPalette.Get("status dimmed"),
		text:     common.DefaultPalette.Get("status text"),
		title:    common.DefaultPalette.Get("status title"),
		success:  common.DefaultPalette.Get("status success"),
		error:    common.DefaultPalette.Get("status error"),
	}

	// Spinner for external commands (legacy, not used for refresh)
	s := spinner.New()
	s.Spinner = spinner.Dot

	t := textinput.New()
	t.Width = 50
	t.TextStyle = styles.text
	t.CompletionStyle = styles.dimmed
	t.PlaceholderStyle = styles.dimmed

	// Spinner characters for single-cell spinner (use spinner.Dot frames)
	spinChars := []rune{'⣾', '⣽', '⣻', '⢿', '⡿', '⣟', '⣯', '⣷'}

	return Model{
		context:      context,
		refreshCount: 0,
		spinnerChars: spinChars,
		spinnerIdx:   0,
		spinner:      s,
		command:      "",
		status:  none,
		input:        t,
		keyMap:       nil,
		styles:       styles,
		history:      make(map[string][]string),
	}
}
