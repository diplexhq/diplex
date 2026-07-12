package handler

import "net/http"

type Handler interface {
	Path() string
	ServeHTTP(http.ResponseWriter, *http.Request)
}

type HTTPServer struct {
	mux      *http.ServeMux
	handlers []Handler
}

func NewHTTPServer(handlers []Handler) *HTTPServer {
	return &HTTPServer{
		mux:      http.NewServeMux(),
		handlers: handlers,
	}
}

func (s *HTTPServer) Handle(mux *http.ServeMux) {
	for _, h := range s.handlers {
		mux.Handle(h.Path(), h)
	}
}
