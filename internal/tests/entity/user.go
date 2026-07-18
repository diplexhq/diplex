package entity

import "time"

type (
	UserAlias = User
	User      struct {
		ID    int
		Name  string
		Email string
	}
)

type UserRepo struct {
	UserStorage
	cache   UserCache
	storage UserStorage
}

type UserStorage interface {
	Get(string) (UserAlias, bool)
	Set(string, UserAlias)
	Delete(string)
}

type UserCache interface {
	Get(string) (UserAlias, bool)
	Set(string, UserAlias, time.Time)
	Delete(string)
}

func NewUserRepo(redisStorage UserStorage, cache UserCache) *UserRepo {
	return &UserRepo{
		cache:   cache,
		storage: redisStorage,
	}
}

func (u *UserRepo) Get(string) (User, bool) {
	return User{}, true
}

func (u *UserRepo) Set(string, User) {}

func (u *UserRepo) Delete(string) {}

func (u *UserRepo) Slug() string {
	return "user"
}

func (u *UserRepo) Stats() int {
	return 42
}
