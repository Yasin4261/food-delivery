package storage

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocal_SaveGeneratesSafeNames(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocal(dir)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	url, err := store.Save(context.Background(), ".jpg", strings.NewReader("data"))
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if !strings.HasPrefix(url, URLPrefix) {
		t.Errorf("url = %q, want %s prefix", url, URLPrefix)
	}
	name := strings.TrimPrefix(url, URLPrefix)
	if !ValidName(name) {
		t.Errorf("generated name %q does not match its own whitelist", name)
	}

	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil || string(data) != "data" {
		t.Errorf("stored content = %q (%v), want %q", data, err, "data")
	}

	// Names are unique across saves.
	url2, _ := store.Save(context.Background(), ".jpg", strings.NewReader("other"))
	if url == url2 {
		t.Error("two saves produced the same name")
	}

	// Unknown extensions are refused.
	if _, err := store.Save(context.Background(), ".php", strings.NewReader("x")); err == nil {
		t.Error("save with .php extension must fail")
	}
}

func TestValidName_RejectsTraversalAndJunk(t *testing.T) {
	bad := []string{
		"../../etc/passwd",
		"..%2f..%2fetc%2fpasswd",
		".env",
		"a.jpg",                          // too short
		strings.Repeat("a", 32) + ".php", // wrong extension
		strings.Repeat("g", 32) + ".jpg", // non-hex
		strings.Repeat("0", 32),          // no extension
	}
	for _, name := range bad {
		if ValidName(name) {
			t.Errorf("ValidName(%q) = true, want false", name)
		}
	}
	if !ValidName(strings.Repeat("0af1", 8) + ".png") {
		t.Error("a well-formed generated name must validate")
	}
}
