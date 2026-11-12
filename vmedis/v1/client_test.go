package vmedisv1

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/time/rate"

	"github.com/turfaa/vmedis-proxy-api/pkg2/retry"
)

// sequenceTokenProvider returns its tokens in order, one per call, repeating
// the last one when it runs out.
type sequenceTokenProvider struct {
	tokens []string
	calls  atomic.Int32
}

func (p *sequenceTokenProvider) GetActiveToken() (string, error) {
	i := int(p.calls.Add(1)) - 1
	return p.tokens[min(i, len(p.tokens)-1)], nil
}

// newRedirectingServer serves "report" for the "good" session and, like
// vmedis does for sessions that cannot open a page, redirects every other
// session to the home page.
func newRedirectingServer(t *testing.T) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			_, _ = io.WriteString(w, "<title>Vmedis - Beranda</title>")
			return
		}

		cookie, err := r.Cookie("vmedisApp")
		if err != nil || cookie.Value != "good" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		_, _ = io.WriteString(w, "report")
	}))
	t.Cleanup(srv.Close)

	return srv
}

func newTestClient(baseUrl string, tokens tokenProvider) *Client {
	c := New(baseUrl, 1, rate.NewLimiter(rate.Inf, 1), tokens)
	c.retryConfig = retry.Config{MaxRetries: 3, InitialBackoff: time.Millisecond}
	return c
}

func TestGet_RetriesRedirectWithAnotherToken(t *testing.T) {
	srv := newRedirectingServer(t)
	tokens := &sequenceTokenProvider{tokens: []string{"bad", "good"}}
	c := newTestClient(srv.URL, tokens)

	res, err := c.get(context.Background(), "/report")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	if string(body) != "report" {
		t.Errorf("got body %q, want %q", body, "report")
	}
	if got := tokens.calls.Load(); got != 2 {
		t.Errorf("got %d token picks, want 2", got)
	}
}

func TestGet_ReturnsErrRedirectedWhenNoTokenCanOpenThePage(t *testing.T) {
	srv := newRedirectingServer(t)
	c := newTestClient(srv.URL, &sequenceTokenProvider{tokens: []string{"bad"}})

	_, err := c.get(context.Background(), "/report")
	if !errors.Is(err, ErrRedirected) {
		t.Fatalf("got error %v, want ErrRedirected", err)
	}
}

func TestGet_RequestingHomePageIsNotARedirect(t *testing.T) {
	srv := newRedirectingServer(t)
	c := newTestClient(srv.URL, &sequenceTokenProvider{tokens: []string{"bad"}})

	res, err := c.get(context.Background(), "/")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	res.Body.Close()
}

func TestGetWithSessionId_AllowedRedirectIsNotAnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/dashboard", http.StatusFound)
			return
		}
		_, _ = io.WriteString(w, "dashboard")
	}))
	t.Cleanup(srv.Close)
	c := newTestClient(srv.URL, &sequenceTokenProvider{tokens: []string{"any"}})

	res, err := c.getWithSessionId(context.Background(), "/", "any", "/dashboard")
	if err != nil {
		t.Fatalf("getWithSessionId: %v", err)
	}
	res.Body.Close()

	if _, err := c.getWithSessionId(context.Background(), "/", "any"); !errors.Is(err, ErrRedirected) {
		t.Fatalf("got error %v without the allowed redirect, want ErrRedirected", err)
	}
}
