package main

import (
	"flag"
	"strings"
)

type Flags struct {
	Path          string
	RulesToIgnore []string
}

func ParseFlags() Flags {
	var path string
	flag.StringVar(&path, "path", "", "path to go templates htmls")

	var rulesToIgnore RulesToIgnoreFlags
	flag.Var(&rulesToIgnore, "ignore-rules", "rules to ignore")

	flag.Parse()

	return Flags{
		Path:          path,
		RulesToIgnore: rulesToIgnore,
	}
}

type RulesToIgnoreFlags []string

func (rs *RulesToIgnoreFlags) String() string {
	if rs == nil {
		return ""
	}
	return strings.Join(*rs, ",")
}

func (rs *RulesToIgnoreFlags) Set(value string) error {
	*rs = append(*rs, value)
	return nil
}
