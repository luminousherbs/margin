package middleware

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"margin.at/internal/logger"
)

func PrivacyLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		t1 := time.Now()

		defer func() {
			safeURL := redactURL(r.URL)

			logger.Info("[%d] %s %s %s",
				ww.Status(),
				r.Method,
				safeURL,
				time.Since(t1),
			)
		}()

		next.ServeHTTP(ww, r)
	})
}

func redactURL(u *url.URL) string {
	redacted := *u
	q := redacted.Query()

	sensitiveKeys := []string{"source", "url", "target", "parent", "root", "uri"}

	for _, key := range sensitiveKeys {
		if q.Has(key) {
			val := q.Get(key)
			if strings.Contains(val, "margin.at") {
				continue
			}
			q.Set(key, "[REDACTED]")
		}
	}

	redacted.RawQuery = q.Encode()
	return redacted.String()
}
