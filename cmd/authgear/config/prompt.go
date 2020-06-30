package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
)

type prompt struct {
	Title        string
	DefaultValue interface{}
	Coerce       func(value string) (interface{}, error)
	Validate     func(value interface{}) error
}

func (p prompt) Prompt() interface{} {
	for {
		fmt.Fprintf(os.Stderr, "%s (default '%v'): ", p.Title, p.DefaultValue)

		var value string
		fmt.Scanln(&value)
		if len(value) == 0 {
			return p.DefaultValue
		}

		cValue, err := p.Coerce(value)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid value: %s\n", err.Error())
			continue
		}
		if p.Validate != nil {
			if err := p.Validate(cValue); err != nil {
				fmt.Fprintf(os.Stderr, "invalid value: %s\n", err.Error())
				continue
			}
		}
		return cValue
	}
}

type promptString struct {
	Title        string
	DefaultValue string
	Validate     func(value string) error
}

func (p promptString) Prompt() string {
	var ps = prompt{
		Title:        p.Title,
		DefaultValue: p.DefaultValue,
		Coerce: func(value string) (interface{}, error) {
			return value, nil
		},
		Validate: func(value interface{}) error {
			if p.Validate != nil {
				return p.Validate(value.(string))
			}
			return nil
		},
	}
	return ps.Prompt().(string)
}

type promptInt struct {
	Title        string
	DefaultValue int
	Validate     func(value int) error
}

func (p promptInt) Prompt() int {
	var ps = prompt{
		Title:        p.Title,
		DefaultValue: p.DefaultValue,
		Coerce: func(value string) (interface{}, error) {
			return strconv.Atoi(value)
		},
		Validate: func(value interface{}) error {
			if p.Validate != nil {
				return p.Validate(value.(int))
			}
			return nil
		},
	}
	return ps.Prompt().(int)
}

type promptURL struct {
	Title        string
	DefaultValue string
	Validate     func(value *url.URL) error
}

func (p promptURL) Prompt() string {
	u, err := url.Parse(p.DefaultValue)
	if err != nil {
		panic(err)
	}

	var ps = prompt{
		Title:        p.Title,
		DefaultValue: u,
		Coerce: func(value string) (interface{}, error) {
			u, err := url.Parse(value)
			if err != nil {
				return nil, err
			}
			if u.Scheme == "" || u.Host == "" {
				return nil, errors.New("URL must be absolute")
			}
			return u, nil
		},
		Validate: func(value interface{}) error {
			if p.Validate != nil {
				return p.Validate(value.(*url.URL))
			}
			return nil
		},
	}
	return ps.Prompt().(*url.URL).String()
}

type promptBool struct {
	Title        string
	DefaultValue bool
}

func (p promptBool) Prompt() bool {
	var ps = prompt{
		Title:        p.Title + " [Y/N]",
		DefaultValue: p.DefaultValue,
		Coerce: func(value string) (interface{}, error) {
			switch value {
			case "Y", "y":
				return true, nil
			case "N", "n":
				return false, nil
			default:
				return nil, errors.New("must enter Y/N")
			}
		},
	}
	return ps.Prompt().(bool)
}
