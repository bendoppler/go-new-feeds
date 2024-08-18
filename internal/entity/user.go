package entity

import "time"

type User struct {
	ID             int
	FirstName      string
	LastName       string
	Username       string
	Password       string
	HashedPassword string
	Salt           string
	Email          string
	Birthday       time.Time
}
