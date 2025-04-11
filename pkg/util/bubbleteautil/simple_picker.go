package bubbleteautil

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	styleSimplePickerItemSelected = lipgloss.NewStyle().Foreground(SemanticInfo).Underline(true)
)

type simplePickerKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	UpWrap   key.Binding
	DownWrap key.Binding
}

var simplePickerDefaultKeyMap = simplePickerKeyMap{
	Up:       key.NewBinding(key.WithKeys("k", "up", "ctrl+p")),
	Down:     key.NewBinding(key.WithKeys("j", "down", "ctrl+n")),
	UpWrap:   key.NewBinding(key.WithKeys("shift+tab")),
	DownWrap: key.NewBinding(key.WithKeys("tab")),
}

type SimplePickerItem struct {
	Label string
	Value string
}

type SimplePicker struct {
	Title  string
	Prompt string
	Items  []SimplePickerItem

	ValidateFunc       func(string) error
	Err                error
	showErrorIfPresent bool

	index     int
	completed bool
	focused   bool
}

var _ tea.Model = SimplePicker{}
var _ Model = SimplePicker{}

func NewSimplePicker(picker SimplePicker) SimplePicker {
	return picker
}

func (m SimplePicker) Init() tea.Cmd {
	return nil
}

func (m SimplePicker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	switch msg := msg.(type) {
	case HideErrorMsg:
		m.showErrorIfPresent = false
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, simplePickerDefaultKeyMap.Down):
			if m.index < len(m.Items)-1 {
				m.index += 1
			}
		case key.Matches(msg, simplePickerDefaultKeyMap.Up):
			if m.index > 0 {
				m.index -= 1
			}
		case key.Matches(msg, simplePickerDefaultKeyMap.DownWrap):
			if m.index < len(m.Items)-1 {
				m.index += 1
			} else {
				m.index = 0
			}
		case key.Matches(msg, simplePickerDefaultKeyMap.UpWrap):
			if m.index > 0 {
				m.index -= 1
			} else {
				m.index = len(m.Items) - 1
			}
		}
	}
	if m.ValidateFunc != nil {
		m.Err = m.ValidateFunc(m.Items[m.index].Value)
	}

	return m, nil
}

func (m SimplePicker) View() string {
	var buf strings.Builder

	fmt.Fprintf(&buf, "%v", m.Title)

	if m.completed {
		fmt.Fprintf(&buf, "%v ", StyleForegroundSemanticSuccess.Render("✓"))
	} else {
		fmt.Fprintf(&buf, "%v ", StyleForegroundSemanticInfo.Render("?"))
	}
	fmt.Fprintf(&buf, "%v:", m.Prompt)

	if m.showErrorIfPresent && m.Err != nil {
		fmt.Fprintf(&buf, " %v", StyleForegroundSemanticError.Render(m.Err.Error()))
	}
	fmt.Fprintf(&buf, "\n")

	for idx, item := range m.Items {
		if idx == m.index {
			fmt.Fprintf(&buf, "%v\n", styleSimplePickerItemSelected.Render(fmt.Sprintf("%v %v", ">", item.Label)))
		} else {
			fmt.Fprintf(&buf, "%v %v\n", " ", item.Label)
		}
	}

	return buf.String()
}

func (m SimplePicker) Value() string {
	return m.Items[m.index].Value
}

func (m SimplePicker) WithValue(val string) Model {
	for idx := range m.Items {
		if m.Items[idx].Value == val {
			m.index = idx
			return m
		}
	}
	panic(fmt.Errorf("value %#v does not match any value of items", val))
}

func (m SimplePicker) Focus() Model {
	m.completed = false
	m.focused = true
	return m
}

func (m SimplePicker) Blur() Model {
	m.completed = true
	m.focused = false
	return m
}

func (m SimplePicker) Validate() (updated Model, valid bool) {
	m.showErrorIfPresent = true
	valid = m.Err == nil
	updated = m
	return
}

func (m SimplePicker) WithError(err error) Model {
	m.showErrorIfPresent = true
	m.Err = err
	return m
}
