package cobraviper

import (
	"fmt"
	"strconv"
)

type IntValue struct {
	Val *int
}

func NewIntValue(val *int) *IntValue {
	return &IntValue{
		Val: val,
	}
}

func (i *IntValue) Set(strVal string) error {
	val, err := strconv.Atoi(strVal)
	if err != nil {
		return fmt.Errorf("invalid integer")
	}
	i.Val = &val
	return nil
}
func (i *IntValue) Type() string { return "int" }

func (i *IntValue) String() string {
	if i.Val != nil {
		return strconv.Itoa(*i.Val)
	}
	return ""
}

type IntArgument struct {
	ArgumentName string
	EnvName      string
	Short        string
	Usage        string
	DefaultValue *int
	Max          *int
	Min          *int
}
