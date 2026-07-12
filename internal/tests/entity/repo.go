package entity

import "errors"

var ErrNotFound = errors.New("entity not found")

// FilterArgs — any для тестирования normalization.
type FilterArgs any

// UserAlias и OrderAlias — alias chain для тестирования nested alias resolution.
type (
	UserAlias  = User
	OrderAlias = Order
)

// DBProvider — узкий интерфейс для тестирования interface narrowing.
type DBProvider interface {
	DSN() string
}

// Repo — generic repository с constraint [T User | Order].
type Repo[T User | Order] struct {
	dsn  string
	slug string
}

func NewRepo[T OrderAlias](conn DBProvider) *Repo[T] {
	return &Repo[T]{dsn: conn.DSN()}
}

func NewYaRepo(conn DBProvider) *Repo[User] {
	return &Repo[User]{dsn: conn.DSN()}
}

func (r *Repo[T]) Get(id int) (T, error)             { var zero T; return zero, ErrNotFound }
func (r *Repo[T]) Create(entity T) (T, error)        { return entity, nil }
func (r *Repo[T]) Update(entity T) (T, error)        { return entity, nil }
func (r *Repo[T]) Delete(e T) error                  { return ErrNotFound }
func (r *Repo[T]) Find(args FilterArgs) ([]T, error) { return nil, nil }
func (r *Repo[T]) Lookup(email string) (T, error)    { var zero T; return zero, ErrNotFound }
func (r *Repo[T]) Detail(id int) (T, error)          { var zero T; return zero, ErrNotFound }
func (r *Repo[T]) Stats() int                        { return 0 }
func (r *Repo[T]) Slug() string                      { return r.slug }

// SluggedRepo — non-generic реализация для handlerAdminStats.
type SluggedRepo struct {
	dsn  string
	slug string
}

func NewSluggedRepo(conn DBProvider) *SluggedRepo {
	return &SluggedRepo{dsn: conn.DSN(), slug: "slugged"}
}

func (r *SluggedRepo) Get(id int) (User, error)             { return User{}, ErrNotFound }
func (r *SluggedRepo) Create(e User) (User, error)          { return e, nil }
func (r *SluggedRepo) Update(e User) (User, error)          { return e, nil }
func (r *SluggedRepo) Delete(e User) error                  { return ErrNotFound }
func (r *SluggedRepo) Find(args FilterArgs) ([]User, error) { return nil, nil }
func (r *SluggedRepo) Lookup(email string) (User, error)    { return User{}, ErrNotFound }
func (r *SluggedRepo) Detail(id int) (User, error)          { return User{}, ErrNotFound }
func (r *SluggedRepo) Stats() int                           { return 0 }
func (r *SluggedRepo) Slug() string                         { return r.slug }
