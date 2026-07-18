package entity

import "time"

type Order struct {
	ID     int
	UserID int
	Amount float64
}

type OrderAlias = Order

type OrderStorage interface {
	Get(int) (Order, bool)
	Set(int, Order)
	Delete(int)
}

type OrderCache interface {
	Get(int) (Order, bool)
	Set(int, Order, time.Time)
	Delete(int)
}

type OrderRepo struct {
	OrderStorage
	cache   OrderCache
	storage OrderStorage
}

func NewOrderRepo(dbStorage OrderStorage, cache OrderCache) *OrderRepo {
	return &OrderRepo{
		cache:   cache,
		storage: dbStorage,
	}
}

func (u *OrderRepo) Get(int) (Order, bool) {
	return Order{}, true
}

func (u *OrderRepo) Set(int, Order) {}

func (u *OrderRepo) Delete(int) {}

func (u *OrderRepo) Slug() string {
	return "user"
}

func (u *OrderRepo) Stats() int {
	return 42
}
