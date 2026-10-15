package middleware

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"testing"
	"time"
)

// SSE responses proxied through Logger must reach the client event by event,
// not only when the upstream finishes. responseRecorder has to pass Flush through
// to the real ResponseWriter for the reverse proxy's FlushInterval = -1 to work.
func TestLoggerStreamsSSEIncrementally(t *testing.T) {
	const upstreamPause = 1500 * time.Millisecond

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Write([]byte("data: first\n\n"))
		w.(http.Flusher).Flush()
		time.Sleep(upstreamPause)
		w.Write([]byte("data: second\n\n"))
	}))
	defer upstream.Close()

	target, _ := url.Parse(upstream.URL)
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.FlushInterval = -1

	gateway := httptest.NewServer(Logger(proxy))
	defer gateway.Close()

	start := time.Now()
	resp, err := http.Get(gateway.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	line, err := bufio.NewReader(resp.Body).ReadString('\n')
	if err != nil {
		t.Fatalf("reading first event: %v", err)
	}
	elapsed := time.Since(start)

	if !strings.HasPrefix(line, "data: first") {
		t.Fatalf("first line = %q, want the first event", line)
	}
	if elapsed >= upstreamPause {
		t.Fatalf("first event arrived after %v, i.e. only once the upstream finished; Logger is buffering the stream", elapsed)
	}
}

func TestLoggerRecordsStatus(t *testing.T) {
	var got *responseRecorder
	h := Logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = w.(*responseRecorder)
		w.WriteHeader(http.StatusTeapot)
	}))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if got.statusCode != http.StatusTeapot {
		t.Fatalf("statusCode = %d, want %d", got.statusCode, http.StatusTeapot)
	}
}
