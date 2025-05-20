package bubbleteautil

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type SingleLineTextInput struct {
	Title        string
	Prompt       string
	ValidateFunc func(string) error
	IsMasked     bool
	Err          error

	model              textinput.Model
	showErrorIfPresent bool
	completed          bool
}

var _ tea.Model = SingleLineTextInput{}
var _ Model = SingleLineTextInput{}

func NewSingleLineTextInput(textInput SingleLineTextInput) SingleLineTextInput {
	model := textinput.New()
	if textInput.ValidateFunc != nil {
		model.Validate = textinput.ValidateFunc(textInput.ValidateFunc)
	}
	if textInput.IsMasked {
		model.EchoMode = textinput.EchoPassword
		model.EchoCharacter = '*'
	}

	textInput.model = model
	return textInput
}

func (m SingleLineTextInput) Init() tea.Cmd {
	return nil
}

func (m SingleLineTextInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case HideErrorMsg:
		m.showErrorIfPresent = false
	}

	var cmd tea.Cmd
	m.model, cmd = m.model.Update(msg)

	var promptBuf strings.Builder
	if m.completed {
		fmt.Fprintf(&promptBuf, "%v ", StyleForegroundSemanticSuccess.Render("✓"))
	} else {
		fmt.Fprintf(&promptBuf, "%v ", StyleForegroundSemanticInfo.Render("?"))
	}
	fmt.Fprintf(&promptBuf, "%v:", m.Prompt)

	if m.showErrorIfPresent && m.getError() != nil {
		fmt.Fprintf(&promptBuf, " %v", StyleForegroundSemanticError.Render(m.getError().Error()))
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

func (m SingleLineTextInput) Validate() (updated Model, valid bool) {
	m.showErrorIfPresent = true

	// Always run ValidateFunc again to ensure Err is up-to-date.
	// When the input is freshly created, its Update method does not run,
	// Err is nil.
	// But if ValidateFunc returns error on empty input, the validation was incorrectly skipped.
	if m.ValidateFunc != nil {
		m.model.Err = m.ValidateFunc(m.model.Value())
	}

	// Intentionally do not consider m.Err because m.Err is external error.
	valid = m.model.Err == nil
	updated = m
	return
}

func (m SingleLineTextInput) Focus() Model {
	m.completed = false
	m.model.Focus()
	return m
}

func (m SingleLineTextInput) Blur() Model {
	m.completed = true
	m.model.Blur()
	return m
}

func (m SingleLineTextInput) IsFocused() bool {
	return m.model.Focused()
}

func (m SingleLineTextInput) Value() string {
	return m.model.Value()
}

func (m SingleLineTextInput) getError() error {
	if m.Err != nil {
		return m.Err
	}
	if m.model.Err != nil {
		return m.model.Err
	}
	return nil
}

func (m SingleLineTextInput) WithError(err error) Model {
	m.Err = err
	m.showErrorIfPresent = true
	return m
}

func (m SingleLineTextInput) WithValue(val string) Model {
	m.model.SetValue(val)
	return m
}
