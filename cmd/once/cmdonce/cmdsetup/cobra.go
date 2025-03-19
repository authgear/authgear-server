package cmdsetup

import (
	"fmt"
	"net/mail"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/pkg/util/bubbleteautil"
)

type MyApplication struct {
	Questions []tea.Model
	controls  []tea.Model
}

var _ tea.Model = MyApplication{}

type msgInit struct{}

func (m MyApplication) Init() tea.Cmd {
	return func() tea.Msg {
		return msgInit{}
	}
}

func (m MyApplication) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case msgInit:
		cmd = m.nextQuestion()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			cmd = m.nextQuestion()
		}
	}

	cmds := make([]tea.Cmd, len(m.controls))
	for idx := range m.controls {
		m.controls[idx], cmds[idx] = m.controls[idx].Update(msg)
	}

	if cmd != nil {
		cmds = append(cmds, tea.Quit)
	}
	return m, tea.Batch(cmds...)
}

func (m *MyApplication) nextQuestion() tea.Cmd {
	shouldQuit := false

	curr := len(m.controls) - 1
	if curr >= 0 {
		valid, updated := bubbleteautil.CheckInput(m.controls[curr])
		m.controls[curr] = updated
		if !valid {
			return nil
		}
	}

	curr += 1

	if curr >= len(m.Questions) {
		shouldQuit = true
	} else {
		question := m.Questions[curr]
		switch question := question.(type) {
		case bubbleteautil.SingleLineTextInput:
			control := bubbleteautil.NewSingleLineTextInput(question)
			control.Focus()
			m.controls = append(m.controls, control)
		case bubbleteautil.SimplePicker:
			control := bubbleteautil.NewSimplePicker(question)
			control.Focus()
			m.controls = append(m.controls, control)
		}
	}

	if shouldQuit {
		return tea.Quit
	}

	return nil
}

func (m MyApplication) View() string {
	var b strings.Builder
	for i := range m.controls {
		fmt.Fprintf(&b, "%v\n", m.controls[i].View())
	}
	return b.String()
}

var CmdSetup = &cobra.Command{
	Use:   "setup",
	Short: "Set up your Authgear ONCE installation.",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := MyApplication{
			Questions: []tea.Model{
				bubbleteautil.SimplePicker{
					Title: `License Agreement

You must accept the license terms to proceed
Authgear ONCE license agreement: https://authgear.com/once/license

`,
					Prompt: "I've read and accept the terms of Authgear ONCE license agreement",
					Items: []bubbleteautil.SimplePickerItem{
						{
							Label: "Yes",
							Value: "true",
						},
						{
							Label: "No",
							Value: "false",
						},
					},
				},
				bubbleteautil.SingleLineTextInput{
					Prompt: "Enter your email",
					Validate: func(input string) error {
						if input == "" {
							return fmt.Errorf("Please enter an email address")
						}

						addr, err := mail.ParseAddress(input)
						if err != nil {
							return fmt.Errorf("Please enter a valid email address")
						}
						if addr.Name != "" {
							return fmt.Errorf("Please enter an email address without name")
						}
						if addr.Address != input {
							return fmt.Errorf("Please enter an email address without spaces")
						}
						return nil
					},
				},
				bubbleteautil.SingleLineTextInput{
					Prompt: "Enter your username",
				},
			},
		}
		prog := tea.NewProgram(app)
		model, err := prog.Run()
		if err != nil {
			return err
		}

		app = model.(MyApplication)
		return nil
	},
}
