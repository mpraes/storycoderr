package fastapi

import "strings"

func joinRoutePath(prefix, route string) string {
	if prefix == "" {
		return ensureLeadingSlash(strings.TrimRight(route, "/"))
	}
	combined := strings.TrimRight(prefix, "/") + "/" + strings.TrimLeft(route, "/")
	return ensureLeadingSlash(strings.TrimRight(combined, "/"))
}

func ensureLeadingSlash(path string) string {
	if path == "" {
		return "/"
	}
	if strings.HasPrefix(path, "/") {
		return path
	}
	return "/" + path
}

func entryPointKey(method, path string) string {
	return "http:" + method + ":" + path
}

func unquotePythonString(raw string) (string, bool) {
	body, ok := stripPythonStringPrefix(raw)
	if !ok {
		return "", false
	}
	return unwrapPythonQuotes(body)
}

func stripPythonStringPrefix(raw string) (string, bool) {
	i := 0
	for i < len(raw) && isPythonStringPrefix(raw[i]) {
		if raw[i] == 'f' || raw[i] == 'F' {
			return "", false
		}
		i++
	}
	return raw[i:], true
}

func isPythonStringPrefix(b byte) bool {
	switch b {
	case 'r', 'R', 'u', 'U', 'b', 'B', 'f', 'F':
		return true
	default:
		return false
	}
}

func unwrapPythonQuotes(body string) (string, bool) {
	switch {
	case len(body) >= 6 && strings.HasPrefix(body, `"""`) && strings.HasSuffix(body, `"""`):
		return body[3 : len(body)-3], true
	case len(body) >= 6 && strings.HasPrefix(body, "'''") && strings.HasSuffix(body, "'''"):
		return body[3 : len(body)-3], true
	case len(body) >= 2 && quotedWith(body, '"'):
		return body[1 : len(body)-1], true
	case len(body) >= 2 && quotedWith(body, '\''):
		return body[1 : len(body)-1], true
	default:
		return "", false
	}
}

func quotedWith(body string, quote byte) bool {
	return body[0] == quote && body[len(body)-1] == quote
}
