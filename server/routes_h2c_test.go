package server

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func TestH2CStreamingParity(t *testing.T) {
	// 1. Setup a Gin router that handles a streaming response
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/generate", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.Stream(func(w io.Writer) bool {
			_, _ = w.Write([]byte(`{"response":"hello"}`))
			return false // stop after one chunk
		})
	})

	// 2. Wrap it in a custom ServeMux to match routes.go configuration
	mux := http.NewServeMux()
	mux.Handle("/", r)

	// 3. Create h2c handler
	h2s := &http2.Server{}
	h2cHandler := h2c.NewHandler(mux, h2s)

	// 4. Start HTTP test server (HTTP cleartext)
	srv := httptest.NewUnstartedServer(h2cHandler)
	srv.Start()
	defer srv.Close()

	// 5. Build an HTTP/2 cleartext client
	client := &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, network, addr)
			},
		},
	}

	// 6. Make request
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/api/generate", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("h2c request failed: %v", err)
	}
	defer resp.Body.Close()

	// 7. Verify we got HTTP/2 and expected response content
	if resp.ProtoMajor != 2 {
		t.Errorf("expected HTTP/2 protocol, got %s", resp.Proto)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	expected := `{"response":"hello"}`
	if string(body) != expected {
		t.Errorf("expected body %q, got %q", expected, string(body))
	}
}
