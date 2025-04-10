package bubbleteautil

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	ANSIBlack   = lipgloss.ANSIColor(0)
	ANSIRed     = lipgloss.ANSIColor(1)
	ANSIGreen   = lipgloss.ANSIColor(2)
	ANSIYellow  = lipgloss.ANSIColor(3)
	ANSIBlue    = lipgloss.ANSIColor(4)
	ANSIMagenta = lipgloss.ANSIColor(5)
	ANSICyan    = lipgloss.ANSIColor(6)
	ANSIWhite   = lipgloss.ANSIColor(7)
)

var (
	SemanticSuccess = ANSIGreen
	SemanticError   = ANSIRed
	SemanticWarning = ANSIYellow
	SemanticInfo    = ANSICyan
)
