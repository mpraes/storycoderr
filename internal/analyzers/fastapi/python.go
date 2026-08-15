package fastapi

import (
	"strings"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

type pythonCST struct {
	parser *tree_sitter.Parser
	tree   *tree_sitter.Tree
}

func parsePython(source []byte) *pythonCST {
	parser := tree_sitter.NewParser()
	language := tree_sitter.NewLanguage(tree_sitter_python.Language())
	if language == nil || parser.SetLanguage(language) != nil {
		parser.Close()
		return nil
	}
	tree := parser.Parse(source, nil)
	if tree == nil {
		parser.Close()
		return nil
	}
	return &pythonCST{parser: parser, tree: tree}
}

func (c *pythonCST) Close() {
	if c.tree != nil {
		c.tree.Close()
	}
	if c.parser != nil {
		c.parser.Close()
	}
}

func (c *pythonCST) Root() *tree_sitter.Node {
	return c.tree.RootNode()
}

func walkNamed(node *tree_sitter.Node, visit func(*tree_sitter.Node)) {
	visit(node)
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child != nil {
			walkNamed(child, visit)
		}
	}
}

func namedChildren(node *tree_sitter.Node) []*tree_sitter.Node {
	count := node.NamedChildCount()
	out := make([]*tree_sitter.Node, 0, count)
	for i := uint(0); i < count; i++ {
		child := node.NamedChild(i)
		if child != nil {
			out = append(out, child)
		}
	}
	return out
}

func literalString(node *tree_sitter.Node, source []byte) (string, bool) {
	if node == nil {
		return "", false
	}
	switch node.Kind() {
	case "string":
		return unquotePythonString(node.Utf8Text(source))
	case "concatenated_string":
		return concatenatedLiteral(node, source)
	default:
		return "", false
	}
}

func concatenatedLiteral(node *tree_sitter.Node, source []byte) (string, bool) {
	var b strings.Builder
	for _, child := range namedChildren(node) {
		part, ok := literalString(child, source)
		if !ok {
			return "", false
		}
		b.WriteString(part)
	}
	return b.String(), true
}
