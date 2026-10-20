package mage

import "testing"

func TestDeriveActivityState(t *testing.T) {
	cases := []struct {
		event string
		want  string
		ok    bool
	}{
		{"session-start", "active", true},
		{"user-prompt-submit", "active", true},
		{"active", "active", true},
		{"stop", "idle", true},
		{"permission-blocked", "blocked", true},
		{"unknown", "", false},
	}
	for _, c := range cases {
		got, ok := DeriveActivityState(c.event, nil)
		if ok != c.ok || string(got) != c.want {
			t.Fatalf("DeriveActivityState(%q) = (%q, %v), want (%q, %v)", c.event, got, ok, c.want, c.ok)
		}
	}
}
