// Package retry retries operations that fail with transient errors,
// using exponential backoff with jitter.
package retry

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"time"
)

// Config controls how Do retries an operation.
type Config struct {
	// MaxRetries is the number of retries after the first attempt fails.
	// With MaxRetries = 3, the operation runs at most 4 times.
	MaxRetries int

	// InitialBackoff is the delay before the first retry.
	// The delay doubles on every subsequent retry.
	InitialBackoff time.Duration

	// MaxBackoff caps the delay between retries. Zero means no cap.
	MaxBackoff time.Duration
}

// DefaultConfig retries 3 times with backoff of 500ms, 1s, and 2s
// (each with up to 50% jitter added).
var DefaultConfig = Config{
	MaxRetries:     100,
	InitialBackoff: 500 * time.Millisecond,
	MaxBackoff:     5 * time.Second,
}

// Do runs op until it succeeds, fails with a permanent error,
// the context is done, or MaxRetries retries are exhausted.
// The last error is returned, annotated with the number of attempts.
func Do[T any](ctx context.Context, cfg Config, op func(ctx context.Context) (T, error)) (T, error) {
	var zero T

	backoff := cfg.InitialBackoff
	for attempt := 1; ; attempt++ {
		result, err := op(ctx)
		if err == nil {
			return result, nil
		}

		if IsPermanent(err) || attempt > cfg.MaxRetries {
			return zero, fmt.Errorf("after %d attempt(s): %w", attempt, err)
		}

		if sleepErr := sleep(ctx, withJitter(backoff)); sleepErr != nil {
			return zero, errors.Join(err, sleepErr)
		}

		backoff *= 2
		if cfg.MaxBackoff > 0 && backoff > cfg.MaxBackoff {
			backoff = cfg.MaxBackoff
		}

		log.Printf("retrying after error: %v (attempt %d/%d, next backoff %s)", err, attempt, cfg.MaxRetries, backoff)
	}
}

// Permanent marks err as non-retryable: Do returns it immediately.
// The original error stays reachable via errors.Is/As.
func Permanent(err error) error {
	if err == nil {
		return nil
	}
	return &permanentError{err: err}
}

// IsPermanent reports whether err is marked with Permanent.
func IsPermanent(err error) bool {
	var pErr *permanentError
	return errors.As(err, &pErr)
}

type permanentError struct {
	err error
}

func (e *permanentError) Error() string { return e.err.Error() }
func (e *permanentError) Unwrap() error { return e.err }

func withJitter(d time.Duration) time.Duration {
	if d <= 0 {
		return d
	}
	return d + rand.N(d/2+1)
}

func sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
