package entity

import "time"

type User struct {
	ID             int
	HashedPassword string
	Salt           string
	FirstName      string
	LastName       string
	Birthday       time.Time
	Email          string
	Username       string // This matches the column `user_name`
	Password       string // Additional field not part of the query
}
