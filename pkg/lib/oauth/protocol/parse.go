package protocol

import "strings"

func parseSpaceDelimitedString(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, " ")
}

func ParseUILocales(s string) []string {
	return parseSpaceDelimitedString(s)
}
