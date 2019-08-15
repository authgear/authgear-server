package command

import (
	"fmt"
	"strings"
)

type Source struct {
	Migration string
	SourceURL string
}

type SourceFlags map[string]*Source

func (i *SourceFlags) String() string {
	return ""
}

func (i *SourceFlags) Set(value string) error {
	s := strings.Split(strings.TrimSpace(value), ",")
	if len(s) != 2 {
		return fmt.Errorf("invalid source format, should be in form <migration>,<src_url> : %s", value)
	}

	migration := s[0]
	url := s[1]

	if _, ok := (*i)[migration]; ok {
		return fmt.Errorf("duplicate source, migration name : %s", migration)
	}

	(*i)[migration] = &Source{
		Migration: migration,
		SourceURL: url,
	}
	return nil
}
