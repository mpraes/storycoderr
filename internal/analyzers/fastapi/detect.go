package fastapi

const (
	FrameworkName    = "fastapi"
	ConfidenceHigh   = "high"
	ConfidenceMedium = "medium"
)

type File struct {
	Path   string
	Source []byte
}

type Route struct {
	Method        string
	Path          string
	HandlerSymbol string
	Framework     string
	EntryPointKey string
	Confidence    string
	Warnings      []string
}

// DetectRoutes finds FastAPI HTTP routes in one Python source file.
// It parses the file with Tree-sitter and never executes Python.
//
//	routes := DetectRoutes(File{Path: "chat.py", Source: src})
func DetectRoutes(file File) []Route {
	cst := parsePython(file.Source)
	if cst == nil {
		return nil
	}
	defer cst.Close()

	root := cst.Root()
	bindings := collectRouterBindings(root, file.Source)
	includes := collectIncludePrefixes(root, file.Source)
	return collectRoutes(file, root, bindings, includes)
}
