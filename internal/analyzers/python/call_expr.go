package python

import tree_sitter "github.com/tree-sitter/go-tree-sitter"

func exprType(node *tree_sitter.Node, file File, env *fileEnv, class *classEnv, locals map[string]string) string {
	if node == nil {
		return ""
	}
	switch node.Kind() {
	case "identifier":
		return identifierType(node.Utf8Text(file.Source), env, locals)
	case "attribute":
		return attributeType(node, file, env, class, locals)
	case "call":
		return callResultType(node, file, env)
	default:
		return ""
	}
}

func identifierType(name string, env *fileEnv, locals map[string]string) string {
	if typ, ok := locals[name]; ok {
		return typ
	}
	if typ, ok := env.names[name]; ok {
		return typ
	}
	return env.resolveTypeName(name)
}

func attributeType(node *tree_sitter.Node, file File, env *fileEnv, class *classEnv, locals map[string]string) string {
	objectType := exprType(node.ChildByFieldName("object"), file, env, class, locals)
	attr := node.ChildByFieldName("attribute")
	if objectType == "" || attr == nil {
		return ""
	}
	owner := env.classByQualified(objectType)
	if owner == nil {
		return ""
	}
	return owner.attrs[attr.Utf8Text(file.Source)]
}

func callResultType(node *tree_sitter.Node, file File, env *fileEnv) string {
	fn := node.ChildByFieldName("function")
	if fn == nil || fn.Kind() != "identifier" {
		return ""
	}
	qualified := env.resolveTypeName(fn.Utf8Text(file.Source))
	if !env.index.isClass(qualified) {
		return ""
	}
	return qualified
}

func paramTypes(def *tree_sitter.Node, file File, env *fileEnv) map[string]string {
	out := map[string]string{}
	params := def.ChildByFieldName("parameters")
	if params == nil {
		return out
	}
	for _, child := range namedChildren(params) {
		name, typeName := parameterTypeName(child, file.Source)
		if name == "" || typeName == "" {
			continue
		}
		if qualified := env.resolveTypeName(typeName); qualified != "" {
			out[name] = qualified
		}
	}
	return out
}

func parameterTypeName(node *tree_sitter.Node, source []byte) (name, typeName string) {
	if node.Kind() != "typed_parameter" && node.Kind() != "typed_default_parameter" {
		return "", ""
	}
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		nameNode = firstKind(node, "identifier")
	}
	if nameNode == nil {
		return "", ""
	}
	return nameNode.Utf8Text(source), typeIdentifier(node.ChildByFieldName("type"), source)
}

func typeIdentifier(node *tree_sitter.Node, source []byte) string {
	if node == nil {
		return ""
	}
	if node.Kind() == "identifier" {
		return node.Utf8Text(source)
	}
	inner := firstKind(node, "identifier")
	if inner == nil {
		return ""
	}
	return inner.Utf8Text(source)
}
