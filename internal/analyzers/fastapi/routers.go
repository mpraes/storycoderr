package fastapi

import (
	"fmt"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type routerBinding struct {
	callee         string
	prefix         string
	prefixResolved bool
}

type includeBinding struct {
	prefix         string
	prefixResolved bool
}

func collectRouterBindings(root *tree_sitter.Node, source []byte) map[string]routerBinding {
	out := map[string]routerBinding{}
	walkNamed(root, func(node *tree_sitter.Node) {
		name, binding, ok := routerBindingFromAssignment(node, source)
		if ok {
			out[name] = binding
		}
	})
	return out
}

func collectIncludePrefixes(root *tree_sitter.Node, source []byte) map[string]includeBinding {
	out := map[string]includeBinding{}
	walkNamed(root, func(node *tree_sitter.Node) {
		name, binding, ok := includeBindingFromCall(node, source)
		if ok {
			out[name] = binding
		}
	})
	return out
}

func routerBindingFromAssignment(assign *tree_sitter.Node, source []byte) (string, routerBinding, bool) {
	if assign.Kind() != "assignment" {
		return "", routerBinding{}, false
	}
	left := assign.ChildByFieldName("left")
	right := assign.ChildByFieldName("right")
	if left == nil || right == nil || left.Kind() != "identifier" || right.Kind() != "call" {
		return "", routerBinding{}, false
	}
	binding, ok := bindingFromConstructor(right, source)
	if !ok {
		return "", routerBinding{}, false
	}
	return left.Utf8Text(source), binding, true
}

func bindingFromConstructor(call *tree_sitter.Node, source []byte) (routerBinding, bool) {
	fn := call.ChildByFieldName("function")
	if fn == nil || fn.Kind() != "identifier" {
		return routerBinding{}, false
	}
	callee := fn.Utf8Text(source)
	if callee != "APIRouter" && callee != "FastAPI" {
		return routerBinding{}, false
	}
	return routerPrefixBinding(callee, call.ChildByFieldName("arguments"), source), true
}

func routerPrefixBinding(callee string, args *tree_sitter.Node, source []byte) routerBinding {
	binding := routerBinding{callee: callee, prefixResolved: true}
	if callee != "APIRouter" || args == nil {
		return binding
	}
	prefix, ok, present := keywordString(args, source, "prefix")
	if !present {
		return binding
	}
	if !ok {
		binding.prefixResolved = false
		return binding
	}
	binding.prefix = prefix
	return binding
}

func includeBindingFromCall(call *tree_sitter.Node, source []byte) (string, includeBinding, bool) {
	if call.Kind() != "call" {
		return "", includeBinding{}, false
	}
	if !isIncludeRouterCall(call, source) {
		return "", includeBinding{}, false
	}
	args := call.ChildByFieldName("arguments")
	name := firstIdentifierArg(args, source)
	if name == "" {
		return "", includeBinding{}, false
	}
	return name, includePrefixBinding(args, source), true
}

func isIncludeRouterCall(call *tree_sitter.Node, source []byte) bool {
	fn := call.ChildByFieldName("function")
	if fn == nil || fn.Kind() != "attribute" {
		return false
	}
	attr := fn.ChildByFieldName("attribute")
	return attr != nil && attr.Utf8Text(source) == "include_router"
}

func includePrefixBinding(args *tree_sitter.Node, source []byte) includeBinding {
	binding := includeBinding{prefixResolved: true}
	if args == nil {
		return binding
	}
	prefix, ok, present := keywordString(args, source, "prefix")
	if !present {
		return binding
	}
	if !ok {
		binding.prefixResolved = false
		return binding
	}
	binding.prefix = prefix
	return binding
}

func firstIdentifierArg(args *tree_sitter.Node, source []byte) string {
	if args == nil {
		return ""
	}
	for _, child := range namedChildren(args) {
		if child.Kind() == "identifier" {
			return child.Utf8Text(source)
		}
	}
	return ""
}

func keywordString(args *tree_sitter.Node, source []byte, name string) (string, bool, bool) {
	for _, child := range namedChildren(args) {
		if child.Kind() != "keyword_argument" {
			continue
		}
		nameNode := child.ChildByFieldName("name")
		if nameNode == nil || nameNode.Utf8Text(source) != name {
			continue
		}
		value, ok := literalString(child.ChildByFieldName("value"), source)
		return value, ok, true
	}
	return "", false, false
}

func resolvePrefix(callee string, bindings map[string]routerBinding, includes map[string]includeBinding) (string, []string, string) {
	prefix, warnings := prefixFromBinding(callee, bindings)
	includePrefix, includeWarnings := prefixFromInclude(callee, includes)
	warnings = append(warnings, includeWarnings...)
	if includePrefix != "" {
		prefix = joinRoutePath(includePrefix, prefix)
	}
	if len(warnings) > 0 {
		return prefix, warnings, ConfidenceMedium
	}
	return prefix, nil, ConfidenceHigh
}

func prefixFromBinding(callee string, bindings map[string]routerBinding) (string, []string) {
	binding, ok := bindings[callee]
	if !ok {
		return "", []string{
			fmt.Sprintf("unresolved router prefix: identifier %q is not assigned to APIRouter(...) or FastAPI() in this file", callee),
		}
	}
	if !binding.prefixResolved {
		return "", []string{
			fmt.Sprintf("unresolved router prefix: %s(%q) prefix is not a string literal, expected prefix=\"/path\"", binding.callee, callee),
		}
	}
	return binding.prefix, nil
}

func prefixFromInclude(callee string, includes map[string]includeBinding) (string, []string) {
	include, ok := includes[callee]
	if !ok {
		return "", nil
	}
	if !include.prefixResolved {
		return "", []string{
			fmt.Sprintf("unresolved include_router prefix: router %q prefix is not a string literal, expected prefix=\"/path\"", callee),
		}
	}
	return include.prefix, nil
}
