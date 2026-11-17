package app

import "testing"

func TestIsDevBrowserOrigin(t *testing.T) {
	t.Parallel()

	cases := map[string]bool{
		"http://localhost:3001":     true,
		"http://127.0.0.1:3001":     true,
		"http://192.168.0.194:3001": true,
		"http://10.0.0.5:3001":      true,
		"http://172.16.4.2:3001":    true,
		"https://localhost:3001":    true,
		"http://example.com":        false,
		"http://172.32.0.1:3001":    false,
		"ftp://localhost:3001":      false,
		"":                          false,
		"not-a-url":                 false,
	}

	for origin, want := range cases {
		if got := isDevBrowserOrigin(origin); got != want {
			t.Fatalf("isDevBrowserOrigin(%q)=%v want %v", origin, got, want)
		}
	}
}

func TestCorsAllowOriginFunc(t *testing.T) {
	t.Parallel()

	fn := corsAllowOriginFunc(false, []string{"http://localhost:3001"})

	origin, allowed, err := fn(nil, "http://127.0.0.1:3001")
	if err != nil || !allowed || origin != "http://127.0.0.1:3001" {
		t.Fatalf("dev localhost variant: origin=%q allowed=%v err=%v", origin, allowed, err)
	}

	origin, allowed, err = fn(nil, "https://evil.example")
	if err != nil || allowed || origin != "" {
		t.Fatalf("blocked foreign origin: origin=%q allowed=%v err=%v", origin, allowed, err)
	}

	prod := corsAllowOriginFunc(true, []string{"http://localhost:3001"})
	origin, allowed, err = prod(nil, "http://127.0.0.1:3001")
	if err != nil || allowed || origin != "" {
		t.Fatalf("prod should not reflect unlisted origin: origin=%q allowed=%v err=%v", origin, allowed, err)
	}
}
