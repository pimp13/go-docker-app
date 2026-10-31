package dto

import "testing"

func TestAuthUserHasAllPermissions(t *testing.T) {
	user := AuthUser{Permissions: []string{"articles:read", "articles:create"}}
	if !user.HasAllPermissions("articles:read") {
		t.Fatal("expected articles:read")
	}
	if user.HasAllPermissions("articles:delete") {
		t.Fatal("did not expect articles:delete")
	}
	if !user.HasAllPermissions("articles:read", "articles:create") {
		t.Fatal("expected both article permissions")
	}
}
