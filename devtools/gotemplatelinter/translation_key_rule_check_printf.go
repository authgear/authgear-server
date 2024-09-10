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
