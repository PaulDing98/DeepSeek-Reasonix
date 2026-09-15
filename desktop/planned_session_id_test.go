package main

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"reasonix/internal/agent"
	"reasonix/internal/session"
)

// plannedFreshSessionIDFromPath is the desktop-side fix for issue
// #10316: it returns the basename (without .jsonl) of the legacy
// transcript that agent.NewSessionPath would mint for a given session
// dir + model. Using that basename as the v4 session id keeps the
// canonical store and the legacy fallback in lockstep.

func TestPlannedFreshSessionIDMatchesLegacyBasename(t *testing.T) {
	dir := t.TempDir()
	// NewSessionPath carries a nanosecond timestamp, so compute it
	// once and pass the result through the helper to avoid the
	// sub-millisecond drift between two back-to-back invocations.
	legacyPath := agent.NewSessionPath(dir, "deepseek-reasonix")
	got := plannedFreshSessionIDFromPath(legacyPath)
	if got == "" {
		t.Fatal("plannedFreshSessionIDFromPath returned empty for a populated path")
	}
	if got != strings.TrimSuffix(filepath.Base(legacyPath), ".jsonl") {
		t.Fatalf("helper must mirror filepath.Base / .jsonl-strip: got %q", got)
	}
	if strings.ContainsAny(got, `/\`) {
		t.Fatalf("planned id %q must not contain a path separator", got)
	}
}

func TestPlannedFreshSessionIDRejectsEmptyPath(t *testing.T) {
	if got := plannedFreshSessionIDFromPath(""); got != "" {
		t.Fatalf("expected empty id for empty path, got %q", got)
	}
}

func TestPlannedFreshSessionIDIsCompatibleWithValidateSessionID(t *testing.T) {
	dir := t.TempDir()
	id := plannedFreshSessionIDFromPath(agent.NewSessionPath(dir, "deepseek-reasonix"))
	// The id we hand to BindFreshSession becomes a v4 directory name
	// via FilesystemPersistence.Create. Verify session.Open accepts
	// it (it runs validateSessionID inside CreateWithOptions). If
	// the id is rejected, the test fails here; success means the
	// identifier is well-formed for the v4 store.
	probe, err := session.Open(filepath.Join(dir, id), id)
	if err != nil {
		t.Fatalf("session.Open rejected planned id %q: %v", id, err)
	}
	t.Cleanup(func() { _ = probe.Close(context.Background()) })
}