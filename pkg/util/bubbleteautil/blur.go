package bubbleteautil

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func CheckInput(model tea.Model) (bool, tea.Model) {
	switch model := model.(type) {
	case SingleLineTextInput:
		valid := model.CheckInput()
		return valid, model
	case SimplePicker:
		valid := model.CheckInput()
		return valid, model
	default:
		panic(fmt.Errorf("%T does not implement CheckInput()", model))
	}
}
