package vmedisv1

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/turfaa/vmedis-proxy-api/database/models"
)

// newVersionedServer acts like vmedis for the sessions "v1" and "v2", which
// can open only their own version's sales page and are redirected to the home
// page from the other one, "none", which is redirected from both, and any
// other session, which gets the login page. It counts the requests it gets.
func newVersionedServer(t *testing.T, requests *atomic.Int32) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)

		cookie, _ := r.Cookie("vmedisApp")
		session := ""
		if cookie != nil {
			session = cookie.Value
		}

		if session != "v1" && session != "v2" && session != "none" {
			_, _ = io.WriteString(w, "<title>Vmedis - Login</title>")
			return
		}

		var pageVersion string
		switch r.URL.Path {
		case "/":
			_, _ = io.WriteString(w, "<title>Vmedis - Beranda</title>")
			return
		case salesPagePaths[models.TokenVersionV1]:
			pageVersion = "v1"
		case salesPagePaths[models.TokenVersionV2]:
			pageVersion = "v2"
		default:
			http.NotFound(w, r)
			return
		}

		if session != pageVersion {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		_, _ = io.WriteString(w, "sales")
	}))
	t.Cleanup(srv.Close)

	return srv
}

func TestRefreshTokens_DetectsVersionAndState(t *testing.T) {
	var requests atomic.Int32
	srv := newVersionedServer(t, &requests)
	c := newTestClient(srv.URL, &sequenceTokenProvider{tokens: []string{"unused"}})

	tokens := []models.VmedisToken{
		{ID: 1, Token: "v1"},
		{ID: 2, Token: "v2"},
		{ID: 3, Token: "none"},
		{ID: 4, Token: "logged-out"},
		{ID: 5, Token: "logged-out-v2", Version: models.TokenVersionV2},
		// A token whose session changed version since it was last checked.
		{ID: 6, Token: "v1", Version: models.TokenVersionV2},
	}

	got, err := c.RefreshTokens(context.Background(), tokens[:5])
	if err != nil {
		t.Fatalf("RefreshTokens: %v", err)
	}

	want := map[string]models.TokenStatus{
		"v1":            {State: models.TokenStateActive, Version: models.TokenVersionV1},
		"v2":            {State: models.TokenStateActive, Version: models.TokenVersionV2},
		"none":          {State: models.TokenStateExpired, Version: models.TokenVersionUnknown},
		"logged-out":    {State: models.TokenStateExpired, Version: models.TokenVersionUnknown},
		"logged-out-v2": {State: models.TokenStateExpired, Version: models.TokenVersionV2},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d statuses, want %d: %+v", len(got), len(want), got)
	}
	for token, w := range want {
		if got[token] != w {
			t.Errorf("token %q: got %+v, want %+v", token, got[token], w)
		}
	}

	changed, err := c.RefreshTokens(context.Background(), tokens[5:])
	if err != nil {
		t.Fatalf("RefreshTokens: %v", err)
	}
	if w := (models.TokenStatus{State: models.TokenStateActive, Version: models.TokenVersionV1}); changed["v1"] != w {
		t.Errorf("token whose version changed: got %+v, want %+v", changed["v1"], w)
	}
}

func TestRefreshTokens_DoesNotRetryRedirects(t *testing.T) {
	var requests atomic.Int32
	srv := newVersionedServer(t, &requests)
	c := newTestClient(srv.URL, &sequenceTokenProvider{tokens: []string{"unused"}})

	if _, err := c.RefreshTokens(context.Background(), []models.VmedisToken{{Token: "v2"}}); err != nil {
		t.Fatalf("RefreshTokens: %v", err)
	}

	// The v1 sales page, its redirect to the home page, then the v2 sales page.
	if got := requests.Load(); got != 3 {
		t.Errorf("got %d requests, want 3", got)
	}
}
