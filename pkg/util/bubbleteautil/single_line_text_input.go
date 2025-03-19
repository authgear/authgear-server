package bubbleteautil

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type SingleLineTextInput struct {
	Title    string
	Prompt   string
	Validate func(string) error

	model            textinput.Model
	dirty            bool
	lastCheckedValue string
	completed        bool
}

var _ tea.Model = SingleLineTextInput{}

func NewSingleLineTextInput(textInput SingleLineTextInput) SingleLineTextInput {
	model := textinput.New()
	if textInput.Validate != nil {
		model.Validate = textinput.ValidateFunc(textInput.Validate)
	}
	textInput.model = model
	return textInput
}

func (m SingleLineTextInput) Init() tea.Cmd {
	return nil
}

func (m SingleLineTextInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.model, cmd = m.model.Update(msg)

	var promptBuf strings.Builder
	if m.completed {
		fmt.Fprintf(&promptBuf, "%v ", StyleForegroundSemanticSuccess.Render("✓"))
	} else {
		fmt.Fprintf(&promptBuf, "%v ", StyleForegroundSemanticInfo.Render("?"))
	}
	fmt.Fprintf(&promptBuf, "%v:", m.Prompt)

	showError := m.Validate != nil && m.dirty && m.lastCheckedValue == m.model.Value() && m.model.Err != nil
	if showError {
		fmt.Fprintf(&promptBuf, " %v", StyleForegroundSemanticError.Render(m.model.Err.Error()))
	}
	fmt.Fprintf(&promptBuf, "\n")

	m.model.Prompt = promptBuf.String()
	return m, cmd
}

func (m SingleLineTextInput) View() string {
	var buf strings.Builder
	fmt.Fprintf(&buf, "%v", m.Title)
	fmt.Fprintf(&buf, "%v", m.model.View())
	if m.completed {
		fmt.Fprintf(&buf, "\n")
	}
	return buf.String()
}

func (m *SingleLineTextInput) CheckInput() (valid bool) {
	m.dirty = true

	if m.Validate != nil {
		m.lastCheckedValue = m.model.Value()
	}
	valid = m.model.Err == nil
	if valid {
		m.completed = true
		m.model.Blur()
	}
	return
}

func (m *SingleLineTextInput) Focus() {
	m.model.Focus()
}

func (m *SingleLineTextInput) Value() string {
	return m.model.Value()
}
