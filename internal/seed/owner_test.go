package seed

import "testing"

func TestSplitOwnerName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in, first, last string
	}{
		{"", "Super", "Admin"},
		{"  ", "Super", "Admin"},
		{"Pouya", "Pouya", "Admin"},
		{"Super Admin", "Super", "Admin"},
		{"Ada Lovelace Byron", "Ada", "Lovelace Byron"},
	}

	for _, tc := range cases {
		first, last := splitOwnerName(tc.in)
		if first != tc.first || last != tc.last {
			t.Fatalf("splitOwnerName(%q)=(%q,%q) want (%q,%q)", tc.in, first, last, tc.first, tc.last)
		}
	}
}
