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
