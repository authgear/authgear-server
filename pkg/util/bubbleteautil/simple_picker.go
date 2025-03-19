package bubbleteautil

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	styleSimplePickerItemSelected = lipgloss.NewStyle().Foreground(SemanticInfo).Underline(true)
)

type SimplePickerItem struct {
	Label string
	Value string
}

type SimplePicker struct {
	Title  string
	Prompt string
	Items  []SimplePickerItem

	index     int
	completed bool
	focused   bool
}

var _ tea.Model = SimplePicker{}

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
	case tea.KeyMsg:
		switch {
		case msg.String() == "j":
			if m.index < len(m.Items)-1 {
				m.index += 1
			}
		case msg.String() == "k":
			if m.index > 0 {
				m.index -= 1
			}
		}
	}

	return m, nil
}

func (m SimplePicker) View() string {
	var buf strings.Builder

	fmt.Fprintf(&buf, "%v", m.Title)

	if m.completed {
		fmt.Fprintf(&buf, "%v %v\n", StyleForegroundSemanticSuccess.Render("✓"), m.Prompt)
	} else {
		fmt.Fprintf(&buf, "%v %v\n", StyleForegroundSemanticInfo.Render("?"), m.Prompt)
	}

	for idx, item := range m.Items {
		if idx == m.index {
			fmt.Fprintf(&buf, "%v\n", styleSimplePickerItemSelected.Render(fmt.Sprintf("%v %v", ">", item.Label)))
		} else {
			fmt.Fprintf(&buf, "%v %v\n", " ", item.Label)
		}
	}

	return buf.String()
}

func (m *SimplePicker) Focus() {
	m.focused = true
}

func (m *SimplePicker) CheckInput() (valid bool) {
	valid = true
	m.completed = true
	m.focused = false
	return
}
