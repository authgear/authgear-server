package cmdsetup

import (
	"fmt"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
)

func validateEmailAddress(input string) error {
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
}

func validateDomain(input string) error {
	if input == "" {
		return fmt.Errorf("Please enter a valid domain")
	}

	re := regexp.MustCompile(`^[-a-zA-Z0-9]+$`)

	parts := strings.Split(input, ".")
	likeInt := make([]bool, len(parts))

	for idx, part := range parts {
		_, err := strconv.Atoi(part)
		likeInt[idx] = err == nil

		if !re.MatchString(part) {
			return fmt.Errorf("Please enter a valid domain")
		}
		if strings.HasPrefix(part, "-") || strings.HasSuffix(part, "-") {
			return fmt.Errorf("Please enter a valid domain")
		}
	}

	allLikeInt := true
	for _, b := range likeInt {
		if !b {
			allLikeInt = false
		}
	}
	if allLikeInt {
		return fmt.Errorf("Please enter a valid domain")
	}
	return nil
}

func validatePassword(input string) error {
	if len(input) < 8 {
		return fmt.Errorf("Please create a password with at least 8 characters")
	}
	return nil
}

func validatePort(input string) error {
	port, err := strconv.Atoi(input)
	if err != nil {
		return fmt.Errorf("Please enter a port in range [1,65535]")
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("Please enter a port in range [1,65535]")
	}
	return nil
}

func validateSMTPUsername(input string) error {
	if input == "" {
		return fmt.Errorf("Please enter an non-empty username")
	}
	return nil
}

func validateSMTPPassword(input string) error {
	if input == "" {
		return fmt.Errorf("Please enter an non-empty password")
	}
	return nil
}

func validateSendgridAPIKey(input string) error {
	if input == "" {
		return fmt.Errorf("Please enter an non-empty Sendgrid API key")
	}
	return nil
}
