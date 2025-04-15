package cliutil

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type Prompt[T any] struct {
	// The title of the prompt.
	// It will be displayed to the end-user.
	Title string

	// The default user input if the end-user skips the prompt.
	// It is a string so that it goes through Parse and Validate.
	InteractiveDefaultUserInput string

	// The flag to read the non-interactive value.
	NonInteractiveFlagName string

	// Parse parses userInput into value.
	Parse func(ctx context.Context, userInput string) (T, error)

	// Validate validates value.
	Validate func(ctx context.Context, value T) error
}

func (p Prompt[T]) Prompt(ctx context.Context, cmd *cobra.Command) (T, error) {
	var zero T

	f := cmd.Flags().Lookup("interactive")
	if f == nil {
		panic(errors.New("programming error: this command has no --interactive defined"))
	}

	isInteractive := false
	interactive := f.Value.String()
	switch interactive {
	case "auto":
		isInteractive = term.IsTerminal(int(os.Stdin.Fd()))
	case "true":
		isInteractive = true
	case "false":
		isInteractive = false
	default:
		return zero, fmt.Errorf("--interactive must be either auto,true,false")
	}

	if !isInteractive {
		f := cmd.Flags().Lookup(p.NonInteractiveFlagName)
		if f == nil {
			panic(fmt.Errorf("programming error: this command has no --%v defined", p.NonInteractiveFlagName))
		}

		var userInput string
		switch {
		// The flag was specified in the command line; Use the specified value.
		case f.Changed:
			userInput = f.Value.String()
		// Otherwise use the default value.
		default:
			userInput = p.InteractiveDefaultUserInput
		}

		parsed, err := p.Parse(ctx, userInput)
		if err != nil {
			return zero, err
		}

		if p.Validate != nil {
			err = p.Validate(ctx, parsed)
			if err != nil {
				return zero, err
			}
		}
		return parsed, nil
	}

	// Otherwise interactive.
	for {
		fmt.Fprintf(os.Stderr, "%v (default '%v'): ", p.Title, p.InteractiveDefaultUserInput)

		var userInput string

		// Deliberately ignore the error because
		// scanning an empty line result in a uncaught-able error.
		_, _ = fmt.Fscanln(os.Stdin, &userInput)
		// Use default if the end-user skips the prompt.
		if len(userInput) == 0 {
			userInput = p.InteractiveDefaultUserInput
		}

		parsed, err := p.Parse(ctx, userInput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid value: %v\n", err)
			continue
		}

		if p.Validate != nil {
			err := p.Validate(ctx, parsed)
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid value: %v\n", err)
				continue
			}
		}

		return parsed, nil
	}
}

type PromptString struct {
	Title                       string
	InteractiveDefaultUserInput string
	NonInteractiveFlagName      string
	Validate                    func(ctx context.Context, value string) error
}

func (p PromptString) Prompt(ctx context.Context, cmd *cobra.Command) (string, error) {
	return Prompt[string]{
		Title:                       p.Title,
		InteractiveDefaultUserInput: p.InteractiveDefaultUserInput,
		NonInteractiveFlagName:      p.NonInteractiveFlagName,
		Parse: func(ctx context.Context, userInput string) (string, error) {
			return userInput, nil
		},
		Validate: p.Validate,
	}.Prompt(ctx, cmd)
}

func (p PromptString) DefineFlag(cmd *cobra.Command) {
	_ = cmd.Flags().String(p.NonInteractiveFlagName, p.InteractiveDefaultUserInput, p.Title)
}

type PromptURL struct {
	Title                       string
	InteractiveDefaultUserInput string
	NonInteractiveFlagName      string
	Validate                    func(ctx context.Context, value *url.URL) error
}

func (p PromptURL) Prompt(ctx context.Context, cmd *cobra.Command) (*url.URL, error) {
	return Prompt[*url.URL]{
		Title:                       p.Title,
		InteractiveDefaultUserInput: p.InteractiveDefaultUserInput,
		NonInteractiveFlagName:      p.NonInteractiveFlagName,
		Parse: func(ctx context.Context, userInput string) (*url.URL, error) {
			u, err := url.Parse(userInput)
			if err != nil {
				return nil, err
			}
			if u.Scheme == "" || u.Host == "" {
				return nil, fmt.Errorf("URL must be absolute: %v", userInput)
			}
			return u, nil
		},
		Validate: p.Validate,
	}.Prompt(ctx, cmd)
}

func (p PromptURL) DefineFlag(cmd *cobra.Command) {
	_ = cmd.Flags().String(p.NonInteractiveFlagName, p.InteractiveDefaultUserInput, p.Title)
}

type PromptBool struct {
	Title                       string
	InteractiveDefaultUserInput bool
	NonInteractiveFlagName      string
}

func (p PromptBool) Prompt(ctx context.Context, cmd *cobra.Command) (bool, error) {
	return Prompt[bool]{
		Title:                       p.Title,
		InteractiveDefaultUserInput: strconv.FormatBool(p.InteractiveDefaultUserInput),
		NonInteractiveFlagName:      p.NonInteractiveFlagName,
		Parse: func(ctx context.Context, userInput string) (bool, error) {
			switch strings.ToLower(userInput) {
			case "y", "yes", "true":
				return true, nil
			case "n", "no", "false":
				return false, nil
			default:
				return false, fmt.Errorf("please enter y/yes/true/n/no/false")
			}
		},
	}.Prompt(ctx, cmd)
}

func (p PromptBool) DefineFlag(cmd *cobra.Command) {
	_ = cmd.Flags().Bool(p.NonInteractiveFlagName, p.InteractiveDefaultUserInput, p.Title)
}

type PromptOptionalPort struct {
	Title                  string
	NonInteractiveFlagName string
}

func (p PromptOptionalPort) Prompt(ctx context.Context, cmd *cobra.Command) (*int, error) {
	return Prompt[*int]{
		Title:                  p.Title,
		NonInteractiveFlagName: p.NonInteractiveFlagName,
		Parse: func(ctx context.Context, userInput string) (*int, error) {
			if userInput == "" {
				return nil, nil
			}
			i, err := strconv.Atoi(userInput)
			if err != nil {
				return nil, fmt.Errorf("Please enter a port within range [1,65535]")
			}
			if i < 1 || i > 65535 {
				return nil, fmt.Errorf("Please enter a port within range [1,65535]")
			}
			return &i, nil
		},
	}.Prompt(ctx, cmd)
}

func (p PromptOptionalPort) DefineFlag(cmd *cobra.Command) {
	_ = cmd.Flags().String(p.NonInteractiveFlagName, "", p.Title)
}

type PromptOptionalEmailAddress struct {
	Title                  string
	NonInteractiveFlagName string
}

func (p PromptOptionalEmailAddress) Prompt(ctx context.Context, cmd *cobra.Command) (string, error) {
	return Prompt[string]{
		Title:                  p.Title,
		NonInteractiveFlagName: p.NonInteractiveFlagName,
		Parse: func(ctx context.Context, userInput string) (string, error) {
			if userInput == "" {
				return "", nil
			}

			addr, err := mail.ParseAddress(userInput)
			if err != nil {
				return "", fmt.Errorf("Please enter a valid email address")
			}
			if addr.Name != "" {
				return "", fmt.Errorf("Please enter an email address without name")
			}
			if addr.Address != userInput {
				return "", fmt.Errorf("Please enter an email address without spaces")
			}
			return userInput, nil
		},
	}.Prompt(ctx, cmd)
}

func (p PromptOptionalEmailAddress) DefineFlag(cmd *cobra.Command) {
	_ = cmd.Flags().String(p.NonInteractiveFlagName, "", p.Title)
}

func DefineFlagInteractive(cmd *cobra.Command) {
	_ = cmd.Flags().String("interactive", "auto", "Run this command interactively: auto,true,false")
}
