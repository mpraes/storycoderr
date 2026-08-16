package python

const (
	KindCalls = "calls"

	ConfidenceHigh   = "high"
	ConfidenceMedium = "medium"
	ConfidenceLow    = "low"
)

type CallRelation struct {
	FromSymbol  string
	ToSymbol    string
	ExternalRef string
	Kind        string
	Line        int
	Confidence  string
}

type CallResult struct {
	Relations []CallRelation
	Warnings  []Warning
}

// ExtractCalls finds direct call relations inside functions and methods.
// It never executes the analyzed files.
//
//	result := ExtractCalls([]File{{Path: "app/api/chat.py", Source: src}})
func ExtractCalls(files []File) CallResult {
	extracted := Extract(files)
	index := newCallIndex(extracted.Symbols)
	result := CallResult{Warnings: extracted.Warnings}
	for _, file := range files {
		extractFileCalls(&result, file, extracted.Symbols, index)
	}
	return result
}

func extractFileCalls(result *CallResult, file File, symbols []Symbol, index *callIndex) {
	cst := parsePython(file.Source)
	if cst == nil {
		return
	}
	defer cst.Close()
	root := cst.Root()
	if root == nil || root.HasError() {
		return
	}
	env := newFileEnv(file.Path, symbols, index)
	collectBindings(root, file, env)
	walkCalls(root, file, moduleName(file.Path), "", env, result)
}
