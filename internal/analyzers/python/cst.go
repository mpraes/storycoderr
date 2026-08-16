package python

import (
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

func identifierText(node *tree_sitter.Node, source []byte) string {
	if node == nil {
		return ""
	}
	name := node.ChildByFieldName("name")
	if name == nil {
		return ""
	}
	return name.Utf8Text(source)
}

func lineRange(node *tree_sitter.Node) (int, int) {
	start := int(node.StartPosition().Row) + 1
	endPos := node.EndPosition()
	end := int(endPos.Row) + 1
	if endPos.Column == 0 && end > start {
		end--
	}
	return start, end
}
