package fastapi

import "testing"

func TestJoinRoutePath(t *testing.T) {
	cases := []struct {
		prefix string
		route  string
		want   string
	}{
		{"/v1", "/chat", "/v1/chat"},
		{"/v1/", "/chat", "/v1/chat"},
		{"/v1", "chat", "/v1/chat"},
		{"", "/v1/chat", "/v1/chat"},
		{"", "chat", "/chat"},
	}
	for _, tc := range cases {
		got := joinRoutePath(tc.prefix, tc.route)
		if got != tc.want {
			t.Fatalf("joinRoutePath(%q, %q) = %q, want %q", tc.prefix, tc.route, got, tc.want)
		}
	}
}

func TestUnquotePythonString(t *testing.T) {
	cases := []struct {
		raw  string
		want string
		ok   bool
	}{
		{`"/v1/chat"`, "/v1/chat", true},
		{`'/v1'`, "/v1", true},
		{`r"/v1"`, "/v1", true},
		{`f"/v1/{x}"`, "", false},
	}
	for _, tc := range cases {
		got, ok := unquotePythonString(tc.raw)
		if ok != tc.ok || got != tc.want {
			t.Fatalf("unquotePythonString(%q) = (%q, %v), want (%q, %v)", tc.raw, got, ok, tc.want, tc.ok)
		}
	}
}

func TestEntryPointKey(t *testing.T) {
	got := entryPointKey("POST", "/v1/chat")
	if got != "http:POST:/v1/chat" {
		t.Fatalf("entryPointKey = %q, want http:POST:/v1/chat", got)
	}
}
