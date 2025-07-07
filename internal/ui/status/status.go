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
)

var cancel = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "dismiss"))
var accept = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "accept"))

type Model struct {
	context      *context.MainContext
	refreshCount int           // Number of refreshes that have occurred
	spinnerChars []rune        // Spinner characters for single-cell spinner
	spinnerIdx   int           // Current spinner index
	showCheck    bool          // Show checkmark after success
	spinner      spinner.Model // Existing spinner for external commands (deprecated for refresh)
	input        textinput.Model
	help         help.Model
	keyMap       help.KeyMap
	command      string
	running      bool
	output       string
	error        error
	width        int
	mode         string
	editing      bool
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
		if string(msg) == "__hide_checkmark__" {
			m.showCheck = false
			return m, nil
		}
		if m.command == string(msg) {
			m.command = ""
			m.error = nil
			m.output = ""
		}
		return m, nil
	case common.CommandRunningMsg:
		m.command = string(msg)
		m.running = true
		return m, m.spinner.Tick
	case common.CommandCompletedMsg:
		m.running = false
		m.output = msg.Output
		m.error = msg.Err
		if m.error == nil {
			commandToBeCleared := m.command
			return m, tea.Tick(CommandClearDuration, func(time.Time) tea.Msg {
				return clearMsg(commandToBeCleared)
			})
		}
		return m, nil
	case common.UpdateRevisionsSuccessMsg:
		// Advance spinner by one tick when refresh is done
		m.refreshCount++
		m.spinnerIdx = m.refreshCount % len(m.spinnerChars)
		m.showCheck = true
		return m, tea.Tick(4*time.Second, func(time.Time) tea.Msg {
			return clearMsg("__hide_checkmark__")
		})
	case common.UpdateRevisionsNoopMsg:
		// Advance spinner by one tick when refresh is noop
		m.refreshCount++
		m.spinnerIdx = m.refreshCount % len(m.spinnerChars)
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, km.Cancel) && m.error != nil:
			m.error = nil
			m.output = ""
			m.command = ""
			m.editing = false
			m.mode = ""
		case key.Matches(msg, km.Cancel) && m.editing:
			m.editing = false
			m.input.Reset()
		case key.Matches(msg, accept) && m.editing:
			editMode := m.mode
			input := m.input.Value()
			prompt := m.input.Prompt

			m.error = nil
			m.output = ""
			m.command = ""
			m.editing = false
			m.mode = ""
			m.input.Reset()
			return m, func() tea.Msg {
				if strings.HasPrefix(editMode, "exec") {
					return exec_process.ExecMsgFromLine(prompt, input)
				}
				return common.QuickSearchMsg(input)
			}
		case key.Matches(msg, km.ExecJJ, km.ExecShell) && !m.editing:
			mode := common.ExecJJ
			if key.Matches(msg, km.ExecShell) {
				mode = common.ExecShell
			}
			m.editing = true
			m.mode = "exec " + mode.Mode
			m.input.Prompt = mode.Prompt
			return m, m.input.Focus()
		case key.Matches(msg, km.QuickSearch) && !m.editing:
			m.editing = true
			m.mode = "search"
			m.input.Prompt = "> "
			return m, m.input.Focus()
		default:
			if m.editing {
				var cmd tea.Cmd
				m.input, cmd = m.input.Update(msg)
				return m, cmd
			}
		}
		return m, nil
	default:
		var cmd tea.Cmd
		// Still update command spinner for legacy/external commands
		if m.running {
			m.spinner, _ = m.spinner.Update(msg)
		}
		return m, cmd
	}
}

func (m *Model) View() string {
	// Single-cell spinner using spinnerChars
	spinnerChar := m.spinnerChars[m.spinnerIdx]
	refreshSpinnerMark := m.styles.title.Render(string(spinnerChar))

	checkMark := m.styles.title.Render(" ")
	if m.showCheck {
		checkMark = m.styles.title.Render("✓")
	}

	commandStatusMark := m.styles.text.Render(" ")
	if m.running {
		commandStatusMark = m.styles.text.Render(m.spinner.View())
	} else if m.error != nil {
		commandStatusMark = m.styles.error.Render("✗ ")
	} else if m.command != "" {
		commandStatusMark = m.styles.success.Render("✓ ")
	} else {
		commandStatusMark = m.help.View(m.keyMap)
	}
	ret := m.styles.text.Render(strings.ReplaceAll(m.command, "\n", "⏎"))
	if m.editing {
		commandStatusMark = ""
		ret = m.input.View()
	}
	mode := m.styles.title.Width(10).Render("", m.mode)
	// Place refresh spinner and checkmark tightly to the left of the mode indicator
	ret = lipgloss.JoinHorizontal(lipgloss.Left, refreshSpinnerMark+checkMark, mode, " ", commandStatusMark, ret)
	if m.error != nil {
		k := cancel.Help().Key
		return lipgloss.JoinVertical(0,
			ret,
			strings.Trim(m.output, "\n"),
			m.styles.shortcut.Render("press ", k, " to dismiss"))
	}
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

	h := help.New()
	h.Styles.ShortKey = styles.shortcut
	h.Styles.ShortDesc = styles.dimmed
	h.Styles.ShortSeparator = styles.dimmed
	h.Styles.FullSeparator = styles.dimmed
	h.Styles.FullKey = styles.shortcut
	h.Styles.FullDesc = styles.dimmed
	h.Styles.Ellipsis = styles.dimmed

	t := textinput.New()
	t.Width = 50

	// Spinner characters for single-cell spinner (use spinner.Dot frames)
	spinChars := []rune{'⣾', '⣽', '⣻', '⢿', '⡿', '⣟', '⣯', '⣷'}

	return Model{
		context:      context,
		refreshCount: 0,
		spinnerChars: spinChars,
		spinnerIdx:   0,
		spinner:      s,
		help:         h,
		command:      "",
		running:      false,
		output:       "",
		input:        t,
		keyMap:       nil,
		styles:       styles,
	}
}
