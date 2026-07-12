// Package logger — узкий интерфейс для тестирования narrow interface per consumer (#12).
package logger

type Logger interface {
	Info(string, ...any)
	Warn(string, ...any)
	Error(string, ...any)
}

type stubLogger struct{}

func (s *stubLogger) Info(msg string, p ...any)  {}
func (s *stubLogger) Warn(msg string, p ...any)  {}
func (s *stubLogger) Error(msg string, p ...any) {}

func NewLogger() Logger {
	return &stubLogger{}
}
