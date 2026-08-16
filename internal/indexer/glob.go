package indexer

import (
	"path"
	"strings"
)

func slashPath(p string) string {
	return strings.ReplaceAll(p, `\`, "/")
}

func matchGlob(pattern, rel string) bool {
	return matchSegs(splitPath(slashPath(pattern)), splitPath(slashPath(rel)))
}

func splitPath(p string) []string {
	p = strings.Trim(p, "/")
	if p == "" {
		return nil
	}
	return strings.Split(p, "/")
}

func matchSegs(pat, segs []string) bool {
	if len(pat) == 0 {
		return len(segs) == 0
	}
	if pat[0] == "**" {
		return matchDoubleStar(pat[1:], segs)
	}
	if len(segs) == 0 {
		return false
	}
	if !matchSegment(pat[0], segs[0]) {
		return false
	}
	return matchSegs(pat[1:], segs[1:])
}

func matchDoubleStar(rest, segs []string) bool {
	for i := 0; i <= len(segs); i++ {
		if matchSegs(rest, segs[i:]) {
			return true
		}
	}
	return false
}

func matchSegment(pattern, name string) bool {
	ok, err := path.Match(pattern, name)
	return err == nil && ok
}
