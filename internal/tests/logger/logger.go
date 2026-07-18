package logger

type Logger interface {
	Info(string, ...any)
	Warn(string, ...any)
	Error(string, ...any)
}

type stubLogger struct{}

func (s *stubLogger) Info(string, ...any)  {}
func (s *stubLogger) Warn(string, ...any)  {}
func (s *stubLogger) Error(string, ...any) {}

func NewLogger() Logger {
	return &stubLogger{}
}
