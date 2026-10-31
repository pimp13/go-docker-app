package rbac

import "testing"

func TestNormalizeRoleSlug(t *testing.T) {
	slug, err := normalizeRoleSlug("  Moderator_1 ")
	if err != nil {
		t.Fatal(err)
	}
	if slug != "moderator_1" {
		t.Fatalf("slug = %q", slug)
	}

	if _, err := normalizeRoleSlug("1bad"); err == nil {
		t.Fatal("expected invalid slug starting with number")
	}
	if _, err := normalizeRoleSlug("Bad Slug"); err == nil {
		t.Fatal("expected invalid slug with spaces")
	}
}

func TestUniqueNonEmpty(t *testing.T) {
	got := uniqueNonEmpty([]string{" articles:read ", "articles:create", "articles:read", ""})
	if len(got) != 2 || got[0] != "articles:read" || got[1] != "articles:create" {
		t.Fatalf("got = %#v", got)
	}
}
