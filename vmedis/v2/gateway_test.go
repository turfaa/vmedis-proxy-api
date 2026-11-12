package vmedisv2

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/turfaa/vmedis-proxy-api/pkg2/retry"
)

const testKey = "test-key"

func TestEncryptToURLEncoded_EncodesSpaceLikeEncodeURIComponent(t *testing.T) {
	c := NewCrypt(testKey)

	// Find a value whose encryption contains a space and a literal "+", the
	// two bytes QueryEscape and encodeURIComponent disagree on.
	var plain string
	for i := 0; ; i++ {
		plain = strings.Repeat("x", i) + `{"method":"POST"}`
		encrypted, _ := c.Encrypt(plain)
		if strings.ContainsRune(string(encrypted), ' ') {
			break
		}
		if i > 100 {
			t.Fatal("no encryption with a space found")
		}
	}

	encoded, err := c.EncryptToURLEncoded(plain)
	if err != nil {
		t.Fatalf("EncryptToURLEncoded: %v", err)
	}

	if strings.Contains(encoded, "+") {
		t.Errorf("encoded %q contains \"+\", which decodeURIComponent does not decode to a space", encoded)
	}

	// decodeURIComponent is equivalent to PathUnescape for these strings.
	decoded, err := url.PathUnescape(encoded)
	if err != nil {
		t.Fatalf("PathUnescape: %v", err)
	}

	var got string
	if err := c.Decrypt([]byte(decoded), &got); err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if got != plain {
		t.Errorf("round trip got %q, want %q", got, plain)
	}

	roundTripped, err := c.DecryptFromURLEncodedToString(encoded)
	if err != nil {
		t.Fatalf("DecryptFromURLEncodedToString: %v", err)
	}
	if roundTripped != plain {
		t.Errorf("DecryptFromURLEncodedToString got %q, want %q", roundTripped, plain)
	}
}

// newTestGateway returns a GatewayClient calling handler, with retries
// disabled.
func newTestGateway(t *testing.T, handler http.HandlerFunc) *GatewayClient {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	c := NewGatewayClient(srv.URL, NewCrypt(testKey))
	c.retryConfig = retry.Config{MaxRetries: 0, InitialBackoff: time.Millisecond}
	return c
}

func encryptForTest(t *testing.T, v any) string {
	t.Helper()

	encrypted, err := NewCrypt(testKey).Encrypt(v)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	return string(encrypted)
}

func TestGatewayCall_SendsEncryptedRequestAndDecryptsResponse(t *testing.T) {
	crypt := NewCrypt(testKey)

	c := newTestGateway(t, func(w http.ResponseWriter, r *http.Request) {
		targetUrl, err := crypt.DecryptFromURLEncodedToString(r.Header.Get("Target-URL"))
		if err != nil || targetUrl != "some/endpoint" {
			t.Errorf("got Target-URL %q (err %v), want %q", targetUrl, err, "some/endpoint")
		}

		options, err := crypt.DecryptFromURLEncodedToString(r.Header.Get("Target-Options"))
		if err != nil || options != `{"method":"POST"}` {
			t.Errorf("got Target-Options %q (err %v)", options, err)
		}

		var body struct{ Params string }
		_ = json.NewDecoder(r.Body).Decode(&body)

		var payload map[string]string
		if err := crypt.Decrypt([]byte(body.Params), &payload); err != nil || payload["q"] != "v" {
			t.Errorf("got payload %v (err %v)", payload, err)
		}

		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "Request processed successfully.",
			"data":    encryptForTest(t, map[string]int{"total": 3}),
		})
	})

	var got map[string]int
	err := c.Call(context.Background(), GatewayRequest{
		TargetUrl:     "some/endpoint",
		TargetVersion: "apivi",
		TargetOptions: GatewayTargetOptions{Method: "POST"},
		Payload:       map[string]string{"q": "v"},
	}, &got)
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if got["total"] != 3 {
		t.Errorf("got %v, want total 3", got)
	}
}

func TestGatewayCall_ReturnsGatewayErrorMessage(t *testing.T) {
	c := newTestGateway(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"message":"Failed to process the request.","error":{"status":0,"message":"Parameter 'login_a' wajib diisi."}}`)
	})

	err := c.Call(context.Background(), GatewayRequest{TargetUrl: "some/endpoint"}, &struct{}{})
	if err == nil || !strings.Contains(err.Error(), "Parameter 'login_a' wajib diisi.") {
		t.Fatalf("got error %v, want it to contain the gateway's message", err)
	}
}

func TestGatewayCall_ReturnsErrorObjectInSuccessfulResponse(t *testing.T) {
	c := newTestGateway(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"error":{"message":"boom"}}`)
	})

	err := c.Call(context.Background(), GatewayRequest{TargetUrl: "some/endpoint"}, &struct{}{})
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("got error %v, want a gateway error containing %q", err, "boom")
	}
}
