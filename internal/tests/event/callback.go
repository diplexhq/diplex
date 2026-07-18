package event

type HandlerFunc func(path string) string

func NewHandlerFunc() HandlerFunc {
	return func(path string) string {
		return "handled:" + path
	}
}
