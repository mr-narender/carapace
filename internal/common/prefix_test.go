package common

import "testing"

func TestPrefixMatcherAdd(t *testing.T) {
	pm := PrefixMatcher{""}

	pm.Add('-')
	if pm.string != "-" {
		t.Errorf(`should be "-" [was: "%v"]`, pm)
	}

	pm.Add('*')
	if pm.string != "*" {
		t.Errorf(`should be "*" [was: "%v"]`, pm)
	}
}

func TestPrefixMatcherMatches(t *testing.T) {
	pm := PrefixMatcher{""}
	pm.Add('-')

	if pm.Matches("") {
		t.Error("empty string should not match")
	}
	if pm.Matches("x") {
		t.Error("string not starting with prefix should not match")
	}
	if !pm.Matches("-x") {
		t.Error("string starting with prefix should match")
	}
	if !pm.Matches("--verbose") {
		t.Error("string starting with prefix should match")
	}
}

func TestPrefixMatcherWildcard(t *testing.T) {
	pm := PrefixMatcher{""}
	pm.Add('*')

	if !pm.Matches("x") {
		t.Error("wildcard should match any string")
	}
	if !pm.Matches("-x") {
		t.Error("wildcard should match any string")
	}
}

func TestPrefixMatcherMerge(t *testing.T) {
	pm1 := PrefixMatcher{""}
	pm1.Add('-')

	pm2 := PrefixMatcher{""}
	pm2.Add('/')

	pm1.Merge(pm2)
	if !pm1.Matches("/path") {
		t.Error("merge should add prefixes from both")
	}
	if !pm1.Matches("-flag") {
		t.Error("merge should keep original prefixes")
	}
}
