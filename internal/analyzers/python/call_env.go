package python

import tree_sitter "github.com/tree-sitter/go-tree-sitter"

type classEnv struct {
	qualified string
	attrs     map[string]string
}

type fileEnv struct {
	index   *callIndex
	imports map[string]string
	names   map[string]string
	classes map[string]*classEnv
}

func newFileEnv(path string, symbols []Symbol, index *callIndex) *fileEnv {
	env := &fileEnv{
		index:   index,
		imports: map[string]string{},
		names:   map[string]string{},
		classes: map[string]*classEnv{},
	}
	for _, symbol := range symbols {
		if symbol.SourceFile != path {
			continue
		}
		bindSymbol(env, symbol)
	}
	return env
}

func bindSymbol(env *fileEnv, symbol Symbol) {
	switch symbol.Kind {
	case KindImport:
		if symbol.DisplayName == "*" {
			return
		}
		env.imports[symbol.DisplayName] = symbol.QualifiedName
	case KindClass:
		env.classes[symbol.DisplayName] = &classEnv{
			qualified: symbol.QualifiedName,
			attrs:     map[string]string{},
		}
	}
}

func (env *fileEnv) resolveTypeName(name string) string {
	if qualified, ok := env.imports[name]; ok {
		return qualified
	}
	if class, ok := env.classes[name]; ok {
		return class.qualified
	}
	return ""
}

func (env *fileEnv) classByQualified(qualified string) *classEnv {
	for _, class := range env.classes {
		if class.qualified == qualified {
			return class
		}
	}
	return nil
}

func collectBindings(root *tree_sitter.Node, file File, env *fileEnv) {
	walkBind(root, file, env, nil)
}

func walkBind(node *tree_sitter.Node, file File, env *fileEnv, class *classEnv) {
	if node == nil {
		return
	}
	switch node.Kind() {
	case "class_definition":
		walkClassBind(node, file, env)
	case "function_definition":
		bindFunction(node, file, env, class)
	case "decorated_definition":
		walkBind(node.ChildByFieldName("definition"), file, env, class)
	case "assignment":
		if class == nil {
			bindAssignment(node, file, env, nil, env.names)
		}
		walkBindChildren(node, file, env, class)
	default:
		walkBindChildren(node, file, env, class)
	}
}

func walkClassBind(def *tree_sitter.Node, file File, env *fileEnv) {
	class := env.classes[identifierText(def, file.Source)]
	for _, child := range namedChildren(def) {
		walkBind(child, file, env, class)
	}
}

func walkBindChildren(node *tree_sitter.Node, file File, env *fileEnv, class *classEnv) {
	for _, child := range namedChildren(node) {
		walkBind(child, file, env, class)
	}
}

func bindFunction(def *tree_sitter.Node, file File, env *fileEnv, class *classEnv) {
	locals := paramTypes(def, file, env)
	if class != nil {
		locals["self"] = class.qualified
	}
	body := def.ChildByFieldName("body")
	collectAssigns(body, file, env, class, locals)
	walkNestedBinds(body, file, env, class)
}

func walkNestedBinds(node *tree_sitter.Node, file File, env *fileEnv, class *classEnv) {
	if node == nil {
		return
	}
	switch node.Kind() {
	case "function_definition", "class_definition", "decorated_definition":
		walkBind(node, file, env, class)
	default:
		for _, child := range namedChildren(node) {
			walkNestedBinds(child, file, env, class)
		}
	}
}

func collectAssigns(node *tree_sitter.Node, file File, env *fileEnv, class *classEnv, locals map[string]string) {
	if node == nil || isDefinition(node) {
		return
	}
	if node.Kind() == "assignment" {
		bindAssignment(node, file, env, class, locals)
	}
	for _, child := range namedChildren(node) {
		collectAssigns(child, file, env, class, locals)
	}
}

func bindAssignment(node *tree_sitter.Node, file File, env *fileEnv, class *classEnv, locals map[string]string) {
	left := node.ChildByFieldName("left")
	right := node.ChildByFieldName("right")
	typ := exprType(right, file, env, class, locals)
	if left == nil || typ == "" {
		return
	}
	if left.Kind() == "identifier" {
		locals[left.Utf8Text(file.Source)] = typ
		return
	}
	bindSelfAttr(left, file, class, typ)
}

func bindSelfAttr(left *tree_sitter.Node, file File, class *classEnv, typ string) {
	if class == nil || left.Kind() != "attribute" {
		return
	}
	object := left.ChildByFieldName("object")
	attr := left.ChildByFieldName("attribute")
	if object == nil || attr == nil || object.Kind() != "identifier" {
		return
	}
	if object.Utf8Text(file.Source) != "self" {
		return
	}
	class.attrs[attr.Utf8Text(file.Source)] = typ
}

func isDefinition(node *tree_sitter.Node) bool {
	switch node.Kind() {
	case "function_definition", "class_definition", "decorated_definition":
		return true
	}
	return false
}
