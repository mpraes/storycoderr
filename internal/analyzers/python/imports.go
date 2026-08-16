package python

import (
	"strings"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

func emitImports(node *tree_sitter.Node, file File, symbols *[]Symbol) {
	if node.Kind() == "import_statement" {
		emitImportList(node, file, "", symbols)
		return
	}
	fromModule := importFromModule(node, file.Source)
	if wildcard := firstKind(node, "wildcard_import"); wildcard != nil {
		name := qualifyImport(fromModule, "*")
		emitSymbol(symbols, file, wildcard, KindImport, name, "*")
		return
	}
	emitImportList(node, file, fromModule, symbols)
}

func emitImportList(node *tree_sitter.Node, file File, fromModule string, symbols *[]Symbol) {
	for i, child := range namedChildren(node) {
		field := node.FieldNameForNamedChild(uint32(i))
		if field == "module_name" {
			continue
		}
		if child.Kind() != "dotted_name" && child.Kind() != "aliased_import" {
			continue
		}
		imported, display := importedName(child, file.Source)
		if imported == "" {
			continue
		}
		emitSymbol(symbols, file, child, KindImport, qualifyImport(fromModule, imported), display)
	}
}

func importedName(node *tree_sitter.Node, source []byte) (imported, display string) {
	if node.Kind() == "dotted_name" {
		text := node.Utf8Text(source)
		return text, text
	}
	name := node.ChildByFieldName("name")
	alias := node.ChildByFieldName("alias")
	if name == nil {
		return "", ""
	}
	imported = name.Utf8Text(source)
	display = imported
	if alias != nil {
		display = alias.Utf8Text(source)
	}
	return imported, display
}

func importFromModule(node *tree_sitter.Node, source []byte) string {
	mod := node.ChildByFieldName("module_name")
	if mod == nil {
		return ""
	}
	return mod.Utf8Text(source)
}

func qualifyImport(fromModule, name string) string {
	if fromModule == "" {
		return name
	}
	if strings.HasSuffix(fromModule, ".") {
		return fromModule + name
	}
	return fromModule + "." + name
}

func firstKind(node *tree_sitter.Node, kind string) *tree_sitter.Node {
	for _, child := range namedChildren(node) {
		if child.Kind() == kind {
			return child
		}
	}
	return nil
}
