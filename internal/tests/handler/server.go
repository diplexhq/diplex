package handler

import (
	"net/http"

	"github.com/diplexhq/diplex/internal/tests/cache"
)

type Handler interface {
	Path() string
	ServeHTTP(http.ResponseWriter, *http.Request)
}

type HTTPServer struct {
	mux      *http.ServeMux
	handlers []Handler
	settings *Settings
	cacheCfg *cache.CacheConfig
}

func NewHTTPServer(settings *Settings, cacheCfg *cache.CacheConfig, handlers []Handler) *HTTPServer {
	return &HTTPServer{
		mux:      http.NewServeMux(),
		handlers: handlers,
		settings: settings,
		cacheCfg: cacheCfg,
	}
}

func (s *HTTPServer) Handle(mux *http.ServeMux) {
	for _, h := range s.handlers {
		mux.Handle(h.Path(), h)
	}
}
