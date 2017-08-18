package main

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/kataras/go-fs"

	"go.uber.org/zap"
)

type HandleFunc func(w http.ResponseWriter, r *http.Request)

type Middleware func(HandleFunc) HandleFunc

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func NewLoggingMiddleware(logger *zap.Logger) Middleware {
	return func(next HandleFunc) HandleFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func(begin time.Time) {
				logger.Info("call", zap.String("path", r.URL.Path), zap.Duration("took", time.Since(begin)))
			}(time.Now())

			next(w, r)
		}
	}
}

func NewCorsMiddleware() Middleware {
	return func(next HandleFunc) HandleFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")

			if r.Method == "OPTIONS" {
				w.Write([]byte("{}"))

				return
			}

			next(w, r)
		}
	}
}

func NewJsonMiddleware() Middleware {
	return func(next HandleFunc) HandleFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")

			next(w, r)
		}
	}
}

func NewGzipMiddleware(gp *fs.GzipPool) Middleware {
	return func(next HandleFunc) HandleFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var wr http.ResponseWriter

			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				wr = w
			} else {
				w.Header().Set("Content-Encoding", "gzip")

				zw := gp.AcquireGzipWriter(w)
				defer gp.ReleaseGzipWriter(zw)

				wr = gzipResponseWriter{Writer: zw, ResponseWriter: w}
			}

			next(wr, r)
		}
	}
}

func NewAuthorisationMiddleware(c Cache, signingFunc func(token *jwt.Token) (interface{}, error)) Middleware {
	return func(next HandleFunc) HandleFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if auth := r.Header.Get("Authorization"); auth != "" {
				token, err := jwt.Parse(auth, signingFunc)
				if err != nil {
					http.Error(w, "unauthorized", http.StatusForbidden)
					return
				}

				if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
					if t, ok := c.Get("token"); ok {
						if t == claims["token"] {
							next(w, r)
							return
						}

						http.Error(w, "unauthorized", http.StatusForbidden)
						return

					}

					http.Error(w, "unauthorized", http.StatusForbidden)
					return

				}

				http.Error(w, "unauthorized", http.StatusForbidden)
				return

			}

			http.Error(w, "unauthorized", http.StatusForbidden)
			return
		}
	}
}

type cacheHandler struct {
	h http.Handler
	c Cache
}

func NewBrowserCacheMiddleware(cache Cache) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return cacheHandler{
			h: next,
			c: cache,
		}
	}
}

func (ch cacheHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if v, ok := ch.c.Get(filepath.Base(r.URL.Path)); ok {
		e := string(v.(string))
		w.Header().Set("Etag", e)
		w.Header().Set("Cache-Control", "max-age=2592000")

		if match := r.Header.Get("If-None-Match"); match != "" {
			if strings.Contains(match, e) {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}
	}

	ch.h.ServeHTTP(w, r)
}

type gzipHandler struct {
	h  http.Handler
	gp *fs.GzipPool
}

func NewFileGzipMiddleware(gp *fs.GzipPool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return gzipHandler{
			h:  next,
			gp: gp,
		}
	}
}

func (gz gzipHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var wr http.ResponseWriter

	if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		wr = w
	} else {
		w.Header().Set("Content-Encoding", "gzip")

		zw := gz.gp.AcquireGzipWriter(w)
		defer gz.gp.ReleaseGzipWriter(zw)

		wr = gzipResponseWriter{Writer: zw, ResponseWriter: w}
	}

	gz.h.ServeHTTP(wr, r)
}
