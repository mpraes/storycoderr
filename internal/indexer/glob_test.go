package indexer

import "testing"

func TestMatchGlob_doubleStarAndWindowsSeparators(t *testing.T) {
	cases := []struct {
		pattern string
		path    string
		want    bool
	}{
		{"**/*.py", "app/api/chat.py", true},
		{"**/*.py", "chat.py", true},
		{"**/*.py", `app\api\chat.py`, true},
		{"**/*.py", "README.md", false},
		{"tests/**/*.py", "tests/test_chat.py", true},
		{"tests/**/*.py", "app/api/chat.py", false},
		{"docs/**/*.md", "docs/guide.md", true},
		{".git/**", ".git/config", true},
		{".venv/**", ".venv/lib/site.py", true},
		{"**/*.py", "app/módulos/café.py", true},
		{"**/*.py", "my app/chat.py", true},
	}
	for _, tc := range cases {
		got := matchGlob(tc.pattern, tc.path)
		if got != tc.want {
			t.Fatalf("matchGlob(%q, %q) = %v, want %v", tc.pattern, tc.path, got, tc.want)
		}
	}
}
