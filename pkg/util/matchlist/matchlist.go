package matchlist

import (
	"fmt"
	"strings"

	"golang.org/x/text/secure/precis"
)

type MatchList struct {
	entries        []string
	foldCase       bool
	stringsContain bool
}

func New(data string, foldCase bool, stringsContain bool) (*MatchList, error) {
	lines := strings.Split(data, "\n")
	entries := []string{}

	var err error
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		if foldCase {
			p := precis.NewFreeform(precis.FoldCase())
			line, err = p.String(line)
			if err != nil {
				return nil, fmt.Errorf("failed to case fold email at line: %d: %w", i+1, err)
			}
		}
		entries = append(entries, line)
	}
	return &MatchList{
		entries:        entries,
		foldCase:       foldCase,
		stringsContain: stringsContain,
	}, nil
}

func (l *MatchList) NumEntries() int {
	return len(l.entries)
}

func (l *MatchList) Matched(value string) (bool, error) {
	v := value
	var err error
	if l.foldCase {
		p := precis.NewFreeform(precis.FoldCase())
		v, err = p.String(value)
		if err != nil {
			return false, fmt.Errorf("failed to case fold email: %w", err)
		}
	}

	var compare func(input string, item string) bool

	if l.stringsContain {
		compare = func(input string, item string) bool {
			return strings.Contains(input, item)
		}
	} else {
		compare = func(input string, item string) bool {
			return input == item
		}
	}

	for _, e := range l.entries {
		if compare(v, e) {
			return true, nil
		}
	}
	return false, nil
}
