// Package httputil provides HTTP helpers shared by the vmedis clients.
package httputil

import (
	"fmt"
	"net/http"

	"github.com/turfaa/vmedis-proxy-api/pkg2/retry"
)

// EnsureSuccess returns an error if the response has an error status code.
// Errors from statuses that won't be fixed by retrying (most 4xx) are marked
// with retry.Permanent.
func EnsureSuccess(res *http.Response) error {
	if res.StatusCode < http.StatusBadRequest {
		return nil
	}

	err := fmt.Errorf("unexpected response status %q", res.Status)
	if !isRetryableStatus(res.StatusCode) {
		return retry.Permanent(err)
	}

	return err
}

func isRetryableStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusRequestTimeout, http.StatusTooManyRequests:
		return true
	}

	return statusCode >= http.StatusInternalServerError
}
