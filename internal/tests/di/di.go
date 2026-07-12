package di

import (
	"github.com/diplexhq/diplex/internal/tests/handler"
)

type DI interface {
	HttpServer() *handler.HTTPServer
}
