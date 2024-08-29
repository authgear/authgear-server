// This file is mostly copied from pkg/util/template/validation.go
package main

import (
	"strconv"
	"strings"
	"text/template/parse"
)

func TraverseTree(tree *parse.Tree, fn func(n parse.Node, depth int) (cont bool)) {
	traverseTreeVisit(tree.Root, 0, fn)
}

func traverseTreeVisitBranch(n *parse.BranchNode, depth int, fn func(n parse.Node, depth int) (cont bool)) (cont bool) {
	if cont = traverseTreeVisit(n.Pipe, depth, fn); !cont {
		return
	}
	if cont = traverseTreeVisit(n.List, depth, fn); !cont {
		return
	}
	if n.ElseList != nil {
		if cont = traverseTreeVisit(n.ElseList, depth, fn); !cont {
			return false
		}
	}
	return
}

func traverseTreeVisit(n parse.Node, depth int, fn func(n parse.Node, depth int) (cont bool)) (cont bool) {
	cont = fn(n, depth)
	if !cont {
		return
	}

	switch n := n.(type) {
	case *parse.PipeNode:
		for _, cmd := range n.Cmds {
			if cont = traverseTreeVisit(cmd, depth, fn); !cont {
				break
			}
		}
	case *parse.CommandNode:
		for _, arg := range n.Args {
			if pipe, ok := arg.(*parse.PipeNode); ok {
				if cont = traverseTreeVisit(pipe, depth+1, fn); !cont {
					break
				}
			}
		}
	case *parse.ActionNode:
		cont = traverseTreeVisit(n.Pipe, depth, fn)
	case *parse.TemplateNode:
		if n.Pipe != nil {
			if cont = traverseTreeVisit(n.Pipe, depth+1, fn); !cont {
				break
			}
		}
	case *parse.TextNode:
		break
	case *parse.IfNode:
		cont = traverseTreeVisitBranch(&n.BranchNode, depth, fn)
	case *parse.RangeNode:
		cont = traverseTreeVisitBranch(&n.BranchNode, depth, fn)
	case *parse.WithNode:
		cont = traverseTreeVisitBranch(&n.BranchNode, depth, fn)
	case *parse.ListNode:
		for _, n := range n.Nodes {
			if cont = traverseTreeVisit(n, depth+1, fn); !cont {
				break
			}
		}
	}

	return
}

func TreeErrorContext(tree *parse.Tree, n parse.Node) (line int, col int, ctx string) {
	location, ctx := tree.ErrorContext(n)
	locationInfo := strings.Split(location, ":") // location is of %s:%d:%d

	line, err := strconv.Atoi(locationInfo[1])
	if err != nil {
		line = 0
	}
	col, err = strconv.Atoi(locationInfo[2])
	if err != nil {
		col = 0
	}

	return
}
