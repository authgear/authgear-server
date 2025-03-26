package template

import (
	"fmt"
	"io"
	"reflect"
	texttemplate "text/template"
	"text/template/parse"
)

// AGTextTemplate is a wrapper of text/template, execpt that <no value> will always be rendered as empty string
type AGTextTemplate struct {
	originalTemplate  *texttemplate.Template
	rewrittenTemplate *texttemplate.Template
}

func (t *AGTextTemplate) Wrap(tpl *texttemplate.Template) error {
	t.originalTemplate = tpl
	return t.rewrite()
}

func (t *AGTextTemplate) Execute(wr io.Writer, data any) error {
	return t.rewrittenTemplate.Execute(wr, data)
}

func (t *AGTextTemplate) rewrite() error {
	clone, err := t.originalTemplate.Clone()
	if err != nil {
		return err
	}
	clone = clone.Funcs(texttemplate.FuncMap{
		"_value_or_empty_string": valueOrEmptyString,
	})

	tpls := clone.Templates()

	for _, tpl := range tpls {
		if tpl.Tree == nil {
			continue
		}
		for _, node := range tpl.Tree.Root.Nodes {
			rewriteNode(node)
		}
	}

	t.rewrittenTemplate = clone
	return nil
}

// For unit test
func (t *AGTextTemplate) String(templateName string) string {
	var result string
	for _, tpl := range t.rewrittenTemplate.Templates() {
		if tpl.Name() == templateName {
			return tpl.Root.String()
		}
	}
	return result
}

func rewriteNode(node parse.Node) {
	rewriteBranch := func(branchNode *parse.BranchNode) {
		if branchNode.List != nil {
			for _, child := range branchNode.List.Nodes {
				rewriteNode(child)
			}
		}
		if branchNode.ElseList != nil {
			for _, child := range branchNode.ElseList.Nodes {
				rewriteNode(child)
			}
		}
	}
	if node == nil {
		return
	}
	switch node := node.(type) {
	case *parse.ActionNode:
		rewriteNode(node.Pipe)
	case *parse.PipeNode:
		if len(node.Cmds) > 0 {
			evalArgsCmd := newIdentCmd("_value_or_empty_string", node.Cmds[0].Position())
			newCmds := []*parse.CommandNode{}
			newCmds = append(newCmds, node.Cmds...)
			newCmds = append(newCmds, evalArgsCmd)
			node.Cmds = newCmds
		}
	case *parse.IfNode:
		rewriteBranch(&node.BranchNode)
	case *parse.BranchNode:
		rewriteBranch(node)
	case *parse.RangeNode:
		rewriteBranch(&node.BranchNode)
	case *parse.ListNode:
		for _, child := range node.Nodes {
			rewriteNode(child)
		}
	default:
		// No need to modify other nodes
	}
}

// newIdentCmd produces a command containing a single identifier node.
// Copied from https://github.com/golang/go/blob/4f11f8ff7db476c534b9b1ad8910dcdd8bbcf022/src/html/template/escape.go#L413C1-L419C2
func newIdentCmd(identifier string, pos parse.Pos) *parse.CommandNode {
	return &parse.CommandNode{
		NodeType: parse.NodeCommand,
		Args:     []parse.Node{parse.NewIdentifier(identifier).SetTree(nil).SetPos(pos)}, // TODO: SetTree.
	}
}

func valueOrEmptyString(args ...any) string {
	newArgs := []any{}
	for _, arg := range args {
		// Skip nils
		// https://github.com/golang/go/blob/4f11f8ff7db476c534b9b1ad8910dcdd8bbcf022/src/html/template/content.go#L174-L179
		newArg := indirectToStringerOrError(arg)
		if newArg == nil {
			continue
		} else {
			newArgs = append(newArgs, newArg)
		}
	}
	return fmt.Sprint(newArgs...)
}

var (
	errorType       = reflect.TypeFor[error]()
	fmtStringerType = reflect.TypeFor[fmt.Stringer]()
)

// indirectToStringerOrError returns the value, after dereferencing as many times
// as necessary to reach the base type (or nil) or an implementation of fmt.Stringer
// or error.
// Copied from https://github.com/golang/go/blob/4f11f8ff7db476c534b9b1ad8910dcdd8bbcf022/src/html/template/content.go#L138
func indirectToStringerOrError(a any) any {
	if a == nil {
		return nil
	}
	v := reflect.ValueOf(a)
	for !v.Type().Implements(fmtStringerType) && !v.Type().Implements(errorType) && v.Kind() == reflect.Pointer && !v.IsNil() {
		v = v.Elem()
	}
	return v.Interface()
}
