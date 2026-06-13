package handlers

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// NewAIProxy forwards /api/v1/ai/* requests to the internal Python AI service.
// The AI service is never exposed to the browser: it lives only on the Docker
// network and is reached through this single, controlled entry point.
func NewAIProxy(target string) gin.HandlerFunc {
	u, err := url.Parse(target)
	if err != nil {
		panic("invalid AI base URL: " + target)
	}

	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, _ error) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":"AI service unavailable"}`))
	}

	return func(c *gin.Context) {
		// Drop the gateway prefix so the AI service sees /estimate, /zones, …
		c.Request.URL.Path = strings.TrimPrefix(c.Request.URL.Path, "/api/v1/ai")
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
