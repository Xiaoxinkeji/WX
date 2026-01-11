
package domain

import "testing"

func TestSource_ParseAndLabel(t *testing.T) {
	s, err := ParseSource("  WEIBO ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != SourceWeibo {
		t.Fatalf("expected weibo, got %q", s)
	}
	if s.Key() != "weibo" {
		t.Fatalf("unexpected key: %q", s.Key())
	}
	if s.Label() != "??" {
		t.Fatalf("unexpected label: %q", s.Label())
	}

	_, err = ParseSource("unknown")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestAllSources_Valid(t *testing.T) {
	all := AllSources()
	if len(all) != 4 {
		t.Fatalf("expected 4 sources, got %d", len(all))
	}
	for _, s := range all {
		if !s.Valid() {
			t.Fatalf("expected valid source: %q", s)
		}
	}
}
