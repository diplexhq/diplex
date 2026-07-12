// Package entity — User и Order для тестирования generic constraints.
package entity

type User struct {
	ID    int
	Name  string
	Email string
}

type Order struct {
	ID     int
	UserID int
	Amount float64
}
