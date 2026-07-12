package patches

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPostgresStoreReturnsStaleSnapshotAfterRefreshFailure(t *testing.T) {
	stale := &patchReadSnapshot{detailBySlug: map[string]PatchDetail{}}
	store := NewPostgresStore(nil, time.Minute)
	store.snapshot = stale
	store.snapshotExpiresAt = time.Now().Add(-time.Minute)
	refreshCount := 0
	store.buildSnapshotFn = func(context.Context) (*patchReadSnapshot, error) {
		refreshCount++
		return nil, errors.New("database unavailable")
	}

	got, err := store.getSnapshot(context.Background())
	if err != nil {
		t.Fatalf("expected stale fallback, got error %v", err)
	}
	if got != stale {
		t.Fatal("expected existing stale snapshot")
	}

	got, err = store.getSnapshot(context.Background())
	if err != nil {
		t.Fatalf("expected stale snapshot during retry delay, got error %v", err)
	}
	if got != stale {
		t.Fatal("expected existing stale snapshot during retry delay")
	}
	if refreshCount != 1 {
		t.Fatalf("expected one refresh attempt during retry delay, got %d", refreshCount)
	}
}

func TestPostgresStorePropagatesCanceledRefresh(t *testing.T) {
	stale := &patchReadSnapshot{detailBySlug: map[string]PatchDetail{}}
	store := NewPostgresStore(nil, time.Minute)
	store.snapshot = stale
	store.snapshotExpiresAt = time.Now().Add(-time.Minute)
	store.buildSnapshotFn = func(ctx context.Context) (*patchReadSnapshot, error) {
		return nil, ctx.Err()
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := store.getSnapshot(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}
