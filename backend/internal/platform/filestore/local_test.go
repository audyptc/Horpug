package filestore

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestLocalRoundTrip(t *testing.T) {
	store, err := NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	key, err := store.Save("documents/dorm", ".pdf", strings.NewReader("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(key, "documents/dorm/") || !strings.HasSuffix(key, ".pdf") {
		t.Fatalf("unexpected key %q", key)
	}

	data, err := store.Read(key)
	if err != nil || !bytes.Equal(data, []byte("hello")) {
		t.Fatalf("Read = %q, %v", data, err)
	}

	if err := store.Delete(key); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Read(key); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Read after delete err = %v, want ErrNotFound", err)
	}
	// Deleting again is not an error.
	if err := store.Delete(key); err != nil {
		t.Fatalf("second Delete: %v", err)
	}
}

func TestLocalRejectsEscapingKeys(t *testing.T) {
	store, err := NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	// A leading slash is not an escape: the key is joined under the root, so
	// "/etc/passwd" just means <root>/etc/passwd.
	for _, key := range []string{"", "../secret", "documents/../../secret"} {
		if _, err := store.Read(key); err == nil || errors.Is(err, ErrNotFound) {
			t.Errorf("Read(%q) err = %v, want an escape error", key, err)
		}
		if err := store.Delete(key); err == nil {
			t.Errorf("Delete(%q) succeeded, want an escape error", key)
		}
	}
}
