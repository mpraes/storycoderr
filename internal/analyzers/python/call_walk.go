package python

import tree_sitter "github.com/tree-sitter/go-tree-sitter"

func walkCalls(node *tree_sitter.Node, file File, module, className string, env *fileEnv, result *CallResult) {
	if node == nil {
		return
	}
	switch node.Kind() {
	case "class_definition":
		name := identifierText(node, file.Source)
		for _, child := range namedChildren(node) {
			walkCalls(child, file, module, name, env, result)
		}
	case "function_definition":
		emitFunctionCalls(node, file, module, className, env, result)
		walkNestedCalls(node.ChildByFieldName("body"), file, module, className, env, result)
	case "decorated_definition":
		walkCalls(node.ChildByFieldName("definition"), file, module, className, env, result)
	default:
		for _, child := range namedChildren(node) {
			walkCalls(child, file, module, className, env, result)
		}
	}
}

func walkNestedCalls(node *tree_sitter.Node, file File, module, className string, env *fileEnv, result *CallResult) {
	if node == nil {
		return
	}
	if isDefinition(node) {
		walkCalls(node, file, module, className, env, result)
		return
	}
	for _, child := range namedChildren(node) {
		walkNestedCalls(child, file, module, className, env, result)
	}
}

func emitFunctionCalls(def *tree_sitter.Node, file File, module, className string, env *fileEnv, result *CallResult) {
	name := identifierText(def, file.Source)
	if name == "" {
		return
	}
	locals := paramTypes(def, file, env)
	class := env.classes[className]
	if class != nil {
		locals["self"] = class.qualified
	}
	from := joinName(module, name)
	if className != "" {
		from = joinName(module, className+"."+name)
	}
	findCalls(def.ChildByFieldName("body"), file, from, env, class, locals, result)
}

func findCalls(node *tree_sitter.Node, file File, from string, env *fileEnv, class *classEnv, locals map[string]string, result *CallResult) {
	if node == nil || isDefinition(node) {
		return
	}
	if node.Kind() == "call" {
		result.Relations = append(result.Relations, emitCall(node, file, from, env, class, locals))
	}
	for _, child := range namedChildren(node) {
		findCalls(child, file, from, env, class, locals, result)
	}
}

func emitCall(node *tree_sitter.Node, file File, from string, env *fileEnv, class *classEnv, locals map[string]string) CallRelation {
	fn := node.ChildByFieldName("function")
	line, _ := lineRange(node)
	callee := ""
	if fn != nil {
		callee = fn.Utf8Text(file.Source)
	}
	to, confidence, resolved := resolveCallee(fn, file, env, class, locals)
	if resolved {
		return CallRelation{
			FromSymbol: from, ToSymbol: to, Kind: KindCalls,
			Line: line, Confidence: confidence,
		}
	}
	return CallRelation{
		FromSymbol: from, ExternalRef: callee, Kind: KindCalls,
		Line: line, Confidence: confidence,
	}
}

func resolveCallee(fn *tree_sitter.Node, file File, env *fileEnv, class *classEnv, locals map[string]string) (string, string, bool) {
	if fn == nil || fn.Kind() != "attribute" {
		return "", ConfidenceLow, false
	}
	methodNode := fn.ChildByFieldName("attribute")
	if methodNode == nil {
		return "", ConfidenceLow, false
	}
	method := methodNode.Utf8Text(file.Source)
	objectType := exprType(fn.ChildByFieldName("object"), file, env, class, locals)
	if objectType == "" {
		return unresolvedAttribute(method, env.index)
	}
	if qualified, ok := env.index.classMethod(objectType, method); ok {
		return qualified, ConfidenceHigh, true
	}
	return "", ConfidenceLow, false
}

func unresolvedAttribute(method string, index *callIndex) (string, string, bool) {
	if index.methodCount(method) > 1 {
		return "", ConfidenceMedium, false
	}
	return "", ConfidenceLow, false
}
