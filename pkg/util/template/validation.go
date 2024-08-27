package template

import (
	"fmt"
	html "html/template"
	"sort"
	text "text/template"
	"text/template/parse"
)

type Validator struct {
	allowRangeNode      bool
	allowTemplateNode   bool
	allowDeclaration    bool
	allowIdentifierNode bool
	maxDepth            int
}

type ValidatorOption func(*Validator)

func AllowRangeNode(b bool) ValidatorOption {
	return func(v *Validator) {
		v.allowRangeNode = b
	}
}

func AllowTemplateNode(b bool) ValidatorOption {
	return func(v *Validator) {
		v.allowTemplateNode = b
	}
}

func AllowDeclaration(b bool) ValidatorOption {
	return func(v *Validator) {
		v.allowDeclaration = b
	}
}

func AllowIdentifierNode(b bool) ValidatorOption {
	return func(v *Validator) {
		v.allowIdentifierNode = b
	}
}

func MaxDepth(d int) ValidatorOption {
	return func(v *Validator) {
		v.maxDepth = d
	}
}

func NewValidator(opts ...ValidatorOption) *Validator {
	v := &Validator{}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

func (v *Validator) ValidateTextTemplate(template *text.Template) error {
	tpls := template.Templates()
	sort.Slice(tpls, func(i, j int) bool {
		return tpls[i].Name() < tpls[j].Name()
	})

	for _, tpl := range tpls {
		if tpl.Tree == nil {
			continue
		}
		if err := v.validateTree(tpl.Tree); err != nil {
			return err
		}
	}
	return nil
}

func (v *Validator) ValidateHTMLTemplate(template *html.Template) error {
	tpls := template.Templates()
	sort.Slice(tpls, func(i, j int) bool {
		return tpls[i].Name() < tpls[j].Name()
	})

	for _, tpl := range tpls {
		if tpl.Tree == nil {
			continue
		}
		if err := v.validateTree(tpl.Tree); err != nil {
			return err
		}
	}
	return nil
}

func (v *Validator) validateTree(tree *parse.Tree) (err error) {
	validateFn := func(n parse.Node, depth int) (cont bool) {
		maxDepth := v.maxDepth
		if maxDepth == 0 {
			maxDepth = 4
		}
		if depth > maxDepth {
			err = fmt.Errorf("%s: template nested too deep", formatLocation(tree, n))
			return
		}
		switch n := n.(type) {
		case *parse.IfNode, *parse.ListNode, *parse.ActionNode, *parse.TextNode:
			break
		case *parse.PipeNode:
			if len(n.Decl) > 0 && !v.allowDeclaration {
				err = fmt.Errorf("%s: declaration is forbidden", formatLocation(tree, n))
			} else if len(n.Cmds) > 1 {
				err = fmt.Errorf("%s: pipeline is forbidden", formatLocation(tree, n))
			}
		case *parse.CommandNode:
			for _, arg := range n.Args {
				if ident, ok := arg.(*parse.IdentifierNode); ok && !v.allowIdentifierNode && !checkIdentifier(ident.Ident) {
					err = fmt.Errorf("%s: forbidden identifier %s", formatLocation(tree, n), ident.Ident)
					break
				}
			}
		case *parse.RangeNode:
			if v.allowRangeNode {
				break
			} else {
				err = fmt.Errorf("%s: forbidden construct %T", formatLocation(tree, n), n)
			}
		case *parse.TemplateNode:
			if v.allowTemplateNode {
				break
			} else {
				err = fmt.Errorf("%s: forbidden construct %T", formatLocation(tree, n), n)
			}
		default:
			err = fmt.Errorf("%s: forbidden construct %T", formatLocation(tree, n), n)
		}

		return err == nil
	}

	traverseTree(tree, validateFn)
	return
}

var badIdentifiers = []string{
	"print",
	"printf",
	"println",
}

func checkIdentifier(id string) bool {
	for _, badID := range badIdentifiers {
		if id == badID {
			return false
		}
	}
	return true
}

func formatLocation(tree *parse.Tree, n parse.Node) string {
	location, _ := tree.ErrorContext(n)
	return location
}

func traverseTree(tree *parse.Tree, fn func(n parse.Node, depth int) (cont bool)) {
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

func traverseTreeVisit(
	n parse.Node,
	depth int,
	fn func(n parse.Node, depth int) (cont bool)) (cont bool) {
	cont = fn(n, depth)
	if !cont {
		return
	}

	switch n := n.(type) {
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
	case *parse.ActionNode:
		cont = traverseTreeVisit(n.Pipe, depth, fn)
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
	case *parse.TemplateNode:
		if n.Pipe != nil {
			if cont = traverseTreeVisit(n.Pipe, depth+1, fn); !cont {
				break
			}
		}

	case *parse.TextNode:
		break
	default:
		panic("unknown node type")
	}
	return
}
