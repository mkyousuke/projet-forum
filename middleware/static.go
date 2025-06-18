package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

func GzipAndCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/static/") {
			ext := filepath.Ext(r.URL.Path)
			switch ext {
			case ".css", ".js", ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg":
				w.Header().Set("Cache-Control", "public, max-age=86400") // 1 jour
			}
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")

		gz := gzip.NewWriter(w)
		defer gz.Close()

		gw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		next.ServeHTTP(gw, r)
	})
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
