package bubbleteautil

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Model is an interface on top of tea.Model.
// A few methods are added to suit our use case.
type Model interface {
	tea.Model
	Validate() (Model, bool)
	Focus() Model
	Blur() Model
	IsFocused() bool
	Value() string
	WithValue(val string) Model
	WithError(err error) Model
}

type HideErrorMsg struct{}

var _ tea.Msg = HideErrorMsg{}

func HideError() tea.Msg {
	return HideErrorMsg{}
}
