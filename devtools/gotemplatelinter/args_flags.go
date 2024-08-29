package main

import (
	"flag"
	"strings"
)

type ArgsFlags struct {
	Paths         []string
	RulesToIgnore []string
}

func ParseArgsFlags() ArgsFlags {
	var rulesToIgnore RulesToIgnoreFlags
	flag.Var(&rulesToIgnore, "ignore-rule", "rules to ignore, repeat this flag to ignore more rules")

	flag.Parse()

	paths := flag.Args()

	return ArgsFlags{
		Paths:         paths,
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
