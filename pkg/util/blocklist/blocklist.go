package blocklist

import (
	"fmt"
	"regexp"
	"strings"
)

type entry struct {
	negated bool
	pattern *regexp.Regexp
}

var regexPattern = regexp.MustCompile("^/(.+)/$")

// Blocklist implements a simple blocklist checking mechanism,
// supporting RegEx and negation.
type Blocklist struct {
	entries []entry
}

func New(data string) (*Blocklist, error) {
	lines := strings.Split(data, "\n")
	var entries []entry
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// Comment lines start with '#'
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Negated patterns start with '!'
		var e entry
		if strings.HasPrefix(line, "!") {
			line = strings.TrimPrefix(line, "!")
			e.negated = true
		}

		match := regexPattern.FindStringSubmatch(line)
		if len(match) > 0 {
			// RegEx pattern
			pattern, err := regexp.Compile(match[1])
			if err != nil {
				return nil, fmt.Errorf("invalid blocklist entry at line %d: %w", i+1, err)
			}
			e.pattern = pattern
		} else {
			// Plain text pattern
			e.pattern = regexp.MustCompile("^" + regexp.QuoteMeta(line) + "$")
		}

		entries = append(entries, e)
	}
	return &Blocklist{entries: entries}, nil
}

func (l *Blocklist) NumEntries() int {
	return len(l.entries)
}

func (l *Blocklist) IsBlocked(value string) bool {
	blocked := false
	for _, e := range l.entries {
		if !e.pattern.MatchString(value) {
			continue
		}

		blocked = !e.negated
	}
	return blocked
}
