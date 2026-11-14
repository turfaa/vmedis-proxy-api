package vmedisv1

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"time"

	"golang.org/x/time/rate"

	"github.com/turfaa/vmedis-proxy-api/pkg2/retry"
	"github.com/turfaa/vmedis-proxy-api/vmedis/internal/httputil"
)

// Client is the main struct for the vmedis client.
type Client struct {
	BaseUrl string

	httpClient  *http.Client
	concurrency int
	limiter     *rate.Limiter
	retryConfig retry.Config

	tokenProvider tokenProvider
}

// New creates a new client.
func New(
	baseUrl string,
	concurrency int,
	limiter *rate.Limiter,
	tokenProvider tokenProvider,
) *Client {
	return &Client{
		BaseUrl:       baseUrl,
		httpClient:    &http.Client{Timeout: time.Minute},
		concurrency:   concurrency,
		limiter:       limiter,
		retryConfig:   retry.DefaultConfig,
		tokenProvider: tokenProvider,
	}
}

// loginPageMarker is present in the response body when vmedis serves the
// login page instead of the requested page, which it does with a success
// status code when the session token is invalid.
const loginPageMarker = "Vmedis - Login"

// ErrInvalidToken is returned when vmedis responds with the login page,
// which means the session token used is invalid.
var ErrInvalidToken = errors.New("invalid session token: vmedis responded with the login page")

// ErrRedirected is returned when vmedis redirects the request to a different
// page, which means the session used cannot open the requested page. It is
// retryable, as a retry may use a session that can.
var ErrRedirected = errors.New("vmedis redirected the request to a different page")

func (c *Client) get(ctx context.Context, path string) (*http.Response, error) {
	return c.getWithRetry(ctx, path, c.tokenProvider.GetActiveToken, true)
}

// getWithSessionId performs a GET request to vmedis with the given session id,
// retrying transient failures. Being redirected to one of allowedRedirectPaths
// is not an error. Being redirected elsewhere is not retried, as the same
// session would be redirected again. See getWithRetry.
func (c *Client) getWithSessionId(ctx context.Context, path, sessionId string, allowedRedirectPaths ...string) (*http.Response, error) {
	return c.getWithRetry(ctx, path, func() (string, error) { return sessionId, nil }, false, allowedRedirectPaths...)
}

// getWithRetry performs a GET request to vmedis, retrying transient failures.
// The session id is picked by getSessionId on every attempt, so a retry can
// use a different session than the attempt that failed.
// Receiving the login page instead of the requested page is an error,
// reported as ErrInvalidToken. Being redirected to another page is an error
// too, reported as ErrRedirected, unless it is to one of allowedRedirectPaths;
// it is retried only if retryRedirects is set.
// The caller owns the response body of a successful request.
func (c *Client) getWithRetry(ctx context.Context, path string, getSessionId func() (string, error), retryRedirects bool, allowedRedirectPaths ...string) (*http.Response, error) {
	res, err := retry.Do(ctx, c.retryConfig, func(ctx context.Context) (*http.Response, error) {
		sessionId, err := getSessionId()
		if err != nil {
			return nil, retry.Permanent(fmt.Errorf("get session id: %w", err))
		}

		res, err := c.doGet(ctx, path, sessionId)
		if err != nil {
			return nil, fmt.Errorf("with session id %s: %w", sessionId, err)
		}

		if err := ensureSessionActive(res); err != nil {
			return nil, fmt.Errorf("with session id %s: %w", sessionId, err)
		}

		if err := ensureNotRedirected(res, allowedRedirectPaths); err != nil {
			err = fmt.Errorf("with session id %s: %w", sessionId, err)
			if !retryRedirects {
				err = retry.Permanent(err)
			}
			return nil, err
		}

		return res, nil
	})
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", path, err)
	}

	return res, nil
}

// ensureNotRedirected returns ErrRedirected if vmedis redirected the request
// to a different page, which it does (to the home page) when the session
// cannot open the requested page, e.g. a v1-menu page with a session that
// has the v2 menu enabled. Redirects to allowedPaths are not errors.
// On error, the body is closed.
func ensureNotRedirected(res *http.Response, allowedPaths []string) error {
	// res.Request is the last request sent. When it was sent because of a
	// redirect, its Response is that redirect, whose Request is the one before.
	final := res.Request
	if final == nil || final.Response == nil {
		return nil
	}

	first := final
	for first.Response != nil && first.Response.Request != nil {
		first = first.Response.Request
	}

	if first.URL.Path == final.URL.Path || slices.Contains(allowedPaths, final.URL.Path) {
		return nil
	}

	res.Body.Close()
	return fmt.Errorf("%w: %s was redirected to %s", ErrRedirected, first.URL.Path, final.URL.Path)
}

// ensureSessionActive returns ErrInvalidToken if the response is the vmedis
// login page. It consumes the response body and, when the session is active,
// restores it so the caller can still read it. On error, the body is closed.
func ensureSessionActive(res *http.Response) error {
	bodyBytes, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	if bytes.Contains(bodyBytes, []byte(loginPageMarker)) {
		// Retrying with the same session token would get the login page
		// again, so fail immediately.
		return retry.Permanent(ErrInvalidToken)
	}

	res.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	return nil
}

func (c *Client) doGet(ctx context.Context, path, sessionId string) (*http.Response, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, retry.Permanent(fmt.Errorf("wait for rate limiter: %w", err))
	}

	finalPath := c.BaseUrl + path
	log.Printf("GET %s", finalPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, finalPath, nil)
	if err != nil {
		return nil, retry.Permanent(fmt.Errorf("create request: %w", err))
	}

	req.Header.Add("Cookie", "vmedisApp="+sessionId)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}

	if err := httputil.EnsureSuccess(res); err != nil {
		res.Body.Close()
		return nil, err
	}

	return res, nil
}
