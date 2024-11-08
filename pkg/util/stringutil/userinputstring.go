package stringutil

import (
	"encoding"
	"strings"
)

type UserInputString struct {
	UnsafeString string
}

func (s UserInputString) TrimSpace() string {
	return strings.TrimSpace(s.UnsafeString)
}

var _ encoding.TextMarshaler = UserInputString{}
var _ encoding.TextUnmarshaler = &UserInputString{}

func (s *UserInputString) UnmarshalText(text []byte) error {
	*s = UserInputString{
		UnsafeString: string(text),
	}
	return nil
}

func (s UserInputString) MarshalText() ([]byte, error) {
	return []byte(s.UnsafeString), nil
}

func NewUserInputString(unsafeString string) UserInputString {
	return UserInputString{UnsafeString: unsafeString}
}
