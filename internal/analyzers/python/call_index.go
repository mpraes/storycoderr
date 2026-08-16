package python

import "strings"

type callIndex struct {
	classes map[string]struct{}
	methods map[string][]string
}

func newCallIndex(symbols []Symbol) *callIndex {
	index := &callIndex{
		classes: make(map[string]struct{}),
		methods: make(map[string][]string),
	}
	for _, symbol := range symbols {
		index.add(symbol)
	}
	return index
}

func (index *callIndex) add(symbol Symbol) {
	switch symbol.Kind {
	case KindClass:
		index.classes[symbol.QualifiedName] = struct{}{}
	case KindFunction, KindMethod:
		name := lastName(symbol.DisplayName)
		index.methods[name] = append(index.methods[name], symbol.QualifiedName)
	}
}

func (index *callIndex) isClass(qualified string) bool {
	_, ok := index.classes[qualified]
	return ok
}

func (index *callIndex) classMethod(typeName, method string) (string, bool) {
	qualified := joinName(typeName, method)
	for _, candidate := range index.methods[method] {
		if candidate == qualified {
			return qualified, true
		}
	}
	return "", false
}

func (index *callIndex) methodCount(method string) int {
	return len(index.methods[method])
}

func lastName(display string) string {
	if i := strings.LastIndex(display, "."); i >= 0 {
		return display[i+1:]
	}
	return display
}
