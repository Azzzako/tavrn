package fuzzy

import "testing"

func TestScore_ExactMatch(t *testing.T) {
	s := Score("deadmau5", "deadmau5")
	if s < 0 {
		t.Error("exact match should score positive")
	}
}

func TestScore_Prefix(t *testing.T) {
	s := Score("dead", "deadmau5")
	if s < 0 {
		t.Error("prefix should match")
	}
}

func TestScore_FuzzyInOrder(t *testing.T) {
	s := Score("dmu", "deadmau5")
	if s < 0 {
		t.Error("fuzzy in-order should match")
	}
}

func TestScore_NoMatch(t *testing.T) {
	s := Score("xyz", "deadmau5")
	if s >= 0 {
		t.Error("no match should return negative")
	}
}

func TestScore_CaseInsensitive(t *testing.T) {
	s := Score("DEAD", "deadmau5")
	if s < 0 {
		t.Error("case insensitive should match")
	}
}

func TestScore_PrefixBetterThanFuzzy(t *testing.T) {
	prefix := Score("dead", "deadmau5")
	fuzzy := Score("dmu5", "deadmau5")
	if prefix <= fuzzy {
		t.Errorf("prefix (%d) should score higher than fuzzy (%d)", prefix, fuzzy)
	}
}

func TestScore_ExactBetterThanPrefix(t *testing.T) {
	exact := Score("deadmau5", "deadmau5")
	prefix := Score("dead", "deadmau5")
	if exact <= prefix {
		t.Errorf("exact (%d) should score higher than prefix (%d)", exact, prefix)
	}
}

func TestScore_ContiguousBetterThanFuzzy(t *testing.T) {
	contig := Score("mau", "deadmau5")
	fuzzy := Score("dmu", "deadmau5")
	if contig <= fuzzy {
		t.Errorf("contiguous (%d) should score higher than fuzzy (%d)", contig, fuzzy)
	}
}

func TestScore_EmptyQuery(t *testing.T) {
	s := Score("", "deadmau5")
	if s < 0 {
		t.Error("empty query should match everything")
	}
}

func TestMatch_FiltersAndRanks(t *testing.T) {
	names := []string{"deadmau5", "synthwave_84", "dj_shadow", "neur0map"}
	got := Match("d", names)
	// Should include deadmau5 and dj_shadow (prefix), maybe neur0map (no d? no.)
	if len(got) < 2 {
		t.Errorf("expected at least 2 matches, got %d: %v", len(got), got)
	}
	if got[0] != "deadmau5" && got[0] != "dj_shadow" {
		t.Errorf("first match should be a prefix match, got %q", got[0])
	}
}

func TestMatch_EmptyQuery(t *testing.T) {
	names := []string{"alice", "bob"}
	got := Match("", names)
	if len(got) != 2 {
		t.Errorf("empty query should return all, got %d", len(got))
	}
}

func TestMatch_NoMatches(t *testing.T) {
	names := []string{"alice", "bob"}
	got := Match("xyz", names)
	if len(got) != 0 {
		t.Errorf("expected no matches, got %v", got)
	}
}

func TestMatch_FuzzyOrdering(t *testing.T) {
	names := []string{"synthwave_84", "deadmau5"}
	got := Match("dmu", names)
	if len(got) != 1 || got[0] != "deadmau5" {
		t.Errorf("expected [deadmau5], got %v", got)
	}
}
