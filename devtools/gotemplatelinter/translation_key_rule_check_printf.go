package main

import (
	"text/template/parse"
)

// validate `printf` command
//
// e.g. (printf "territory-%s" .Value)
func CheckCommandPrintf(printfNode *parse.CommandNode) (err error) {
	// 2nd arg should be translation key
	for idx, arg := range printfNode.Args {
		if idx == 1 {
			err = CheckTranslationKeyNode(arg)
			if err != nil {
				return err
			}
		}

	}
	return
}

// check if pipe node is only `printf`
//
// e.g. (printf "territory-%s" .Value)
func IsPipeNodeOnlyPrintfCommand(p *parse.PipeNode) bool {
	if p == nil {
		return false
	}
	if len(p.Cmds) == 0 || len(p.Cmds) > 1 {
		return false
	}
	c := p.Cmds[0]
	if c == nil {
		return false
	}
	if len(c.Args) == 0 {
		return false
	}
	ident, ok := c.Args[0].(*parse.IdentifierNode)
	if !ok {
		panic("unexpected command node, first arg not identifier")
	}
	return ident.String() == "printf"
}
