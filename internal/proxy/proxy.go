package proxy

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/basecase/gateway/internal/config"
)

func NewReverseProxy(routes []config.Route) http.Handler {
	mux := http.NewServeMux()

	for _, route := range routes {
		target, err := url.Parse(route.TargetURL)
		if err != nil {
			slog.Error("Invalid target URL", "url", route.TargetURL, "error", err)
			continue
		}

		proxy := httputil.NewSingleHostReverseProxy(target)

		// Customize the Director to modify request before forwarding
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			// Ensure we don't forward host from the original request
			req.Host = target.Host
		}

		// Customize ErrorHandler to return JSON instead of plain text on 502 Bad Gateway
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			slog.Error("Proxy error", "target", target.String(), "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(`{"error": "Bad Gateway"}`))
		}

		// Register the route in the mux
		// The trailing slash makes it act as a prefix match
		path := route.PathPrefix
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}
		
		mux.Handle(path, http.StripPrefix(route.PathPrefix, proxy))
		
		// Also handle exact match without trailing slash
		if !strings.HasSuffix(route.PathPrefix, "/") {
			mux.Handle(route.PathPrefix, proxy)
		}
	}

	// Add a healthcheck endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	return mux
}
