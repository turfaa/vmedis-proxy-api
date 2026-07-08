package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

var testConfig = Config{
	MaxRetries:     3,
	InitialBackoff: time.Millisecond,
}

func TestDo_SucceedsFirstAttempt(t *testing.T) {
	calls := 0
	result, err := Do(context.Background(), testConfig, func(context.Context) (string, error) {
		calls++
		return "ok", nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if result != "ok" {
		t.Fatalf("expected result %q, got %q", "ok", result)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestDo_SucceedsAfterRetries(t *testing.T) {
	calls := 0
	result, err := Do(context.Background(), testConfig, func(context.Context) (int, error) {
		calls++
		if calls < 3 {
			return 0, errors.New("transient")
		}
		return 42, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if result != 42 {
		t.Fatalf("expected result 42, got %d", result)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_ExhaustsRetries(t *testing.T) {
	transient := errors.New("transient")

	calls := 0
	_, err := Do(context.Background(), testConfig, func(context.Context) (int, error) {
		calls++
		return 0, transient
	})

	if !errors.Is(err, transient) {
		t.Fatalf("expected error to wrap %q, got %q", transient, err)
	}
	if wantCalls := testConfig.MaxRetries + 1; calls != wantCalls {
		t.Fatalf("expected %d calls, got %d", wantCalls, calls)
	}
}

func TestDo_StopsOnPermanentError(t *testing.T) {
	permanent := errors.New("permanent")

	calls := 0
	_, err := Do(context.Background(), testConfig, func(context.Context) (int, error) {
		calls++
		return 0, Permanent(permanent)
	})

	if !errors.Is(err, permanent) {
		t.Fatalf("expected error to wrap %q, got %q", permanent, err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestDo_StopsOnContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	calls := 0
	_, err := Do(ctx, testConfig, func(context.Context) (int, error) {
		calls++
		cancel()
		return 0, errors.New("transient")
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected error to wrap context.Canceled, got %q", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestIsPermanent(t *testing.T) {
	if IsPermanent(errors.New("transient")) {
		t.Fatal("expected transient error to not be permanent")
	}

	wrapped := Permanent(errors.New("permanent"))
	if !IsPermanent(wrapped) {
		t.Fatal("expected wrapped error to be permanent")
	}
}
