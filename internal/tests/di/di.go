package di

import (
	"github.com/diplexhq/diplex/internal/tests/handler"
	"github.com/diplexhq/diplex/internal/tests/logger"
)

type DI interface {
	HttpServer() *handler.HTTPServer
	Logger() logger.Logger
}
