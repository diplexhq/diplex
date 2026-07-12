// Package callback — named function type для тестирования function type dependency (#29).
package callback

type HandlerFunc func(path string) string

func NewHandlerFunc() HandlerFunc {
	return func(path string) string {
		return "handled:" + path
	}
}
