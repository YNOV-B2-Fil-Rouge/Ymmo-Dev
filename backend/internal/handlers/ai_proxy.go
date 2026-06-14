package handlers

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

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

	proxy.ModifyResponse = func(resp *http.Response) error {
		for _, h := range []string{
			"Access-Control-Allow-Origin",
			"Access-Control-Allow-Credentials",
			"Access-Control-Allow-Methods",
			"Access-Control-Allow-Headers",
			"Access-Control-Expose-Headers",
			"Access-Control-Max-Age",
		} {
			resp.Header.Del(h)
		}
		return nil
	}

	return func(c *gin.Context) {
		c.Request.URL.Path = strings.TrimPrefix(c.Request.URL.Path, "/api/v1/ai")
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
