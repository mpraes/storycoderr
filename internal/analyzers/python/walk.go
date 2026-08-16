package python

import tree_sitter "github.com/tree-sitter/go-tree-sitter"

func walk(node *tree_sitter.Node, file File, module, className string, symbols *[]Symbol) {
	if node == nil {
		return
	}
	switch node.Kind() {
	case "module":
		emitModule(node, file, module, symbols)
		walkChildren(node, file, module, className, symbols)
	case "class_definition":
		walkClass(node, node, file, module, symbols)
	case "function_definition":
		emitCallable(node, node, file, module, className, symbols)
	case "decorated_definition":
		walkDecorated(node, file, module, className, symbols)
	case "import_statement", "import_from_statement":
		emitImports(node, file, symbols)
	default:
		walkChildren(node, file, module, className, symbols)
	}
}

func walkChildren(node *tree_sitter.Node, file File, module, className string, symbols *[]Symbol) {
	for _, child := range namedChildren(node) {
		walk(child, file, module, className, symbols)
	}
}

func walkClass(def, span *tree_sitter.Node, file File, module string, symbols *[]Symbol) {
	name := identifierText(def, file.Source)
	if name == "" {
		return
	}
	emitSymbol(symbols, file, span, KindClass, joinName(module, name), name)
	walkChildren(def, file, module, name, symbols)
}

func walkDecorated(node *tree_sitter.Node, file File, module, className string, symbols *[]Symbol) {
	for _, child := range namedChildren(node) {
		if child.Kind() == "decorator" {
			emitDecorator(child, file, module, className, symbols)
		}
	}
	def := node.ChildByFieldName("definition")
	if def == nil {
		return
	}
	if def.Kind() == "class_definition" {
		walkClass(def, node, file, module, symbols)
		return
	}
	if def.Kind() == "function_definition" {
		emitCallable(def, node, file, module, className, symbols)
	}
}

func emitModule(node *tree_sitter.Node, file File, module string, symbols *[]Symbol) {
	emitSymbol(symbols, file, node, KindModule, module, moduleDisplay(module))
}

func emitCallable(def, span *tree_sitter.Node, file File, module, className string, symbols *[]Symbol) {
	name := identifierText(def, file.Source)
	if name == "" {
		return
	}
	kind := KindFunction
	display := name
	qualified := joinName(module, name)
	if className != "" {
		kind = KindMethod
		display = className + "." + name
		qualified = joinName(module, display)
	}
	emitSymbol(symbols, file, span, kind, qualified, display)
}

func emitDecorator(node *tree_sitter.Node, file File, module, className string, symbols *[]Symbol) {
	display := decoratorDisplay(node, file.Source)
	if display == "" {
		return
	}
	qualified := joinName(module, display)
	if className != "" {
		qualified = joinName(module, className+"."+display)
	}
	emitSymbol(symbols, file, node, KindDecorator, qualified, display)
}

func emitSymbol(symbols *[]Symbol, file File, node *tree_sitter.Node, kind Kind, qualified, display string) {
	start, end := lineRange(node)
	*symbols = append(*symbols, Symbol{
		QualifiedName: qualified,
		DisplayName:   display,
		Kind:          kind,
		SourceFile:    file.Path,
		StartLine:     start,
		EndLine:       end,
		SemanticHash:  semanticHash(kind, qualified, node.Utf8Text(file.Source)),
	})
}

func joinName(left, right string) string {
	if left == "" {
		return right
	}
	if right == "" {
		return left
	}
	return left + "." + right
}

func decoratorDisplay(node *tree_sitter.Node, source []byte) string {
	inner := node.NamedChild(0)
	if inner == nil {
		return ""
	}
	if inner.Kind() != "call" {
		return inner.Utf8Text(source)
	}
	fn := inner.ChildByFieldName("function")
	if fn == nil {
		return inner.Utf8Text(source)
	}
	return fn.Utf8Text(source)
}
