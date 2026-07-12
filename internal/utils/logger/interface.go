package logger

// Logger defines the contract for logging in diplex.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}
