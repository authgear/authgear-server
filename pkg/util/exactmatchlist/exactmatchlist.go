package exactmatchlist

import (
	"fmt"
	"strings"

	"golang.org/x/text/secure/precis"
)

type ExactMatchList struct {
	entries  []string
	foldCase bool
}

func New(data string, foldCase bool) (*ExactMatchList, error) {
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
	return &ExactMatchList{
		entries:  entries,
		foldCase: foldCase,
	}, nil
}

func (l *ExactMatchList) NumEntries() int {
	return len(l.entries)
}

func (l *ExactMatchList) Matched(value string) (bool, error) {
	v := value
	var err error
	if l.foldCase {
		p := precis.NewFreeform(precis.FoldCase())
		v, err = p.String(value)
		if err != nil {
			return false, fmt.Errorf("failed to case fold email: %w", err)
		}
	}

	for _, e := range l.entries {
		if e == v {
			return true, nil
		}
	}
	return false, nil
}
