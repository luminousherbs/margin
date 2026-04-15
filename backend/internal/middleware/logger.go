package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"margin.at/internal/logger"
)

func PrivacyLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		t1 := time.Now()

		defer func() {
			logger.Info("[%d] %s %s %s",
				ww.Status(),
				r.Method,
				r.URL.String(),
				time.Since(t1),
			)
		}()

		next.ServeHTTP(ww, r)
	})
}
