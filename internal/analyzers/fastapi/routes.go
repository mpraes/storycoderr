package fastapi

import tree_sitter "github.com/tree-sitter/go-tree-sitter"

var httpMethods = map[string]string{
	"get":    "GET",
	"post":   "POST",
	"put":    "PUT",
	"patch":  "PATCH",
	"delete": "DELETE",
}

func collectRoutes(file File, root *tree_sitter.Node, bindings map[string]routerBinding, includes map[string]includeBinding) []Route {
	var routes []Route
	walkNamed(root, func(node *tree_sitter.Node) {
		if node.Kind() != "decorated_definition" {
			return
		}
		route, ok := routeFromDecorated(node, file.Source, bindings, includes)
		if ok {
			routes = append(routes, route)
		}
	})
	return routes
}

func routeFromDecorated(node *tree_sitter.Node, source []byte, bindings map[string]routerBinding, includes map[string]includeBinding) (Route, bool) {
	handler := handlerName(node, source)
	if handler == "" {
		return Route{}, false
	}
	for _, child := range namedChildren(node) {
		if child.Kind() != "decorator" {
			continue
		}
		method, callee, path, ok := httpDecorator(child, source)
		if !ok {
			continue
		}
		return buildRoute(method, path, handler, callee, bindings, includes)
	}
	return Route{}, false
}

func handlerName(decorated *tree_sitter.Node, source []byte) string {
	def := decorated.ChildByFieldName("definition")
	if def == nil || def.Kind() != "function_definition" {
		return ""
	}
	name := def.ChildByFieldName("name")
	if name == nil {
		return ""
	}
	return name.Utf8Text(source)
}

func httpDecorator(decorator *tree_sitter.Node, source []byte) (method, callee, path string, ok bool) {
	call := decorator.NamedChild(0)
	if call == nil || call.Kind() != "call" {
		return "", "", "", false
	}
	method, callee, ok = httpAttribute(call.ChildByFieldName("function"), source)
	if !ok {
		return "", "", "", false
	}
	path, ok = routePath(call.ChildByFieldName("arguments"), source)
	return method, callee, path, ok
}

func httpAttribute(fn *tree_sitter.Node, source []byte) (method, callee string, ok bool) {
	if fn == nil || fn.Kind() != "attribute" {
		return "", "", false
	}
	object := fn.ChildByFieldName("object")
	attr := fn.ChildByFieldName("attribute")
	if object == nil || attr == nil || object.Kind() != "identifier" {
		return "", "", false
	}
	method, ok = httpMethods[attr.Utf8Text(source)]
	if !ok {
		return "", "", false
	}
	return method, object.Utf8Text(source), true
}

func routePath(args *tree_sitter.Node, source []byte) (string, bool) {
	if args == nil {
		return "", false
	}
	if value, ok, present := keywordString(args, source, "path"); present {
		return value, ok
	}
	return firstPositionalString(args, source)
}

func firstPositionalString(args *tree_sitter.Node, source []byte) (string, bool) {
	for _, child := range namedChildren(args) {
		if child.Kind() == "keyword_argument" {
			continue
		}
		return literalString(child, source)
	}
	return "", false
}

func buildRoute(method, decoratorPath, handler, callee string, bindings map[string]routerBinding, includes map[string]includeBinding) (Route, bool) {
	prefix, warnings, confidence := resolvePrefix(callee, bindings, includes)
	path := joinRoutePath(prefix, decoratorPath)
	return Route{
		Method:        method,
		Path:          path,
		HandlerSymbol: handler,
		Framework:     FrameworkName,
		EntryPointKey: entryPointKey(method, path),
		Confidence:    confidence,
		Warnings:      warnings,
	}, true
}
