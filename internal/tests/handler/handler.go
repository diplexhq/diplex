// Package handler — base type shared by all HTTP handlers для тестирования struct embedding (#14).
package handler

type Base struct {
	path string
}

func NewBase(path string) Base {
	return Base{path: path}
}

func (b *Base) Path() string {
	return b.path
}
