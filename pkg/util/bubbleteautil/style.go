package bubbleteautil

import (
	"github.com/charmbracelet/lipgloss"
)

// Common styles.
// For complex ones, please define yourselves.
var (
	StyleForegroundSemanticInfo    = lipgloss.NewStyle().Foreground(SemanticInfo)
	StyleForegroundSemanticSuccess = lipgloss.NewStyle().Foreground(SemanticSuccess)
	StyleForegroundSemanticError   = lipgloss.NewStyle().Foreground(SemanticError)
)
