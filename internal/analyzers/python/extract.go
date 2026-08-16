package python

import (
	"crypto/sha256"
	"encoding/hex"
	"path"
	"strings"
)

type Kind string

const (
	KindModule    Kind = "module"
	KindClass     Kind = "class"
	KindFunction  Kind = "function"
	KindMethod    Kind = "method"
	KindDecorator Kind = "decorator"
	KindImport    Kind = "import"
)

type File struct {
	Path   string
	Source []byte
}

type Symbol struct {
	QualifiedName string
	DisplayName   string
	Kind          Kind
	SourceFile    string
	StartLine     int
	EndLine       int
	SemanticHash  string
}

type Warning struct {
	Path    string
	Message string
}

type Result struct {
	Symbols  []Symbol
	Warnings []Warning
}

// Extract finds Python symbols in source files using Tree-sitter.
// It never executes the analyzed files.
//
//	result := Extract([]File{{Path: "app/api/chat.py", Source: src}})
func Extract(files []File) Result {
	var result Result
	for _, file := range files {
		extractFile(&result, file)
	}
	return result
}

func extractFile(result *Result, file File) {
	cst := parsePython(file.Source)
	if cst == nil {
		result.Warnings = append(result.Warnings, parseWarning(file.Path, "parser unavailable"))
		return
	}
	defer cst.Close()

	root := cst.Root()
	if root == nil || root.HasError() {
		result.Warnings = append(result.Warnings, parseWarning(file.Path, "syntax error"))
		return
	}
	walk(root, file, moduleName(file.Path), "", &result.Symbols)
}

func parseWarning(filePath, reason string) Warning {
	return Warning{
		Path: filePath,
		Message: "cannot extract symbols from Python file " + quote(filePath) +
			" (" + reason + "), expected a parseable module",
	}
}

func quote(value string) string {
	return `"` + value + `"`
}

func moduleName(filePath string) string {
	normalized := strings.ReplaceAll(filePath, "\\", "/")
	normalized = strings.TrimSuffix(normalized, ".py")
	normalized = strings.TrimSuffix(normalized, "/__init__")
	return strings.ReplaceAll(normalized, "/", ".")
}

func moduleDisplay(module string) string {
	base := path.Base(strings.ReplaceAll(module, ".", "/"))
	if base == "." || base == "/" || base == "" {
		return module
	}
	return base
}

func semanticHash(kind Kind, qualifiedName, body string) string {
	sum := sha256.Sum256([]byte(string(kind) + "\x00" + qualifiedName + "\x00" + body))
	return hex.EncodeToString(sum[:])
}
