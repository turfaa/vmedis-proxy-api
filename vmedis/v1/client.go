package vmedisv1

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
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

func (c *Client) get(ctx context.Context, path string) (*http.Response, error) {
	sessionId, err := c.tokenProvider.GetActiveToken()
	if err != nil {
		return nil, fmt.Errorf("get active session id: %w", err)
	}

	return c.getWithSessionId(ctx, path, sessionId)
}

// getWithSessionId performs a GET request to vmedis, retrying transient
// failures. Receiving the login page instead of the requested page is an
// error too, reported as ErrInvalidToken.
// The caller owns the response body of a successful request.
func (c *Client) getWithSessionId(ctx context.Context, path, sessionId string) (*http.Response, error) {
	res, err := retry.Do(ctx, c.retryConfig, func(ctx context.Context) (*http.Response, error) {
		res, err := c.doGet(ctx, path, sessionId)
		if err != nil {
			return nil, err
		}

		if err := ensureSessionActive(res); err != nil {
			return nil, err
		}

		return res, nil
	})
	if err != nil {
		return nil, fmt.Errorf("GET %s with session id %s: %w", path, sessionId, err)
	}

	return res, nil
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
