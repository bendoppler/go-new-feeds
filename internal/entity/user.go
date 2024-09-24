package entity

import (
	"fmt"
	"time"
)

type User struct {
	ID             int       `json:"id"`
	HashedPassword string    `json:"hashed_password"`
	Salt           string    `json:"salt"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	Birthday       time.Time `json:"birthday"`
	Email          string    `json:"email"`
	Username       string    `json:"username"` // This matches the column `user_name`
	Password       string    `json:"password"` // Additional field not part of the query
}

func (u User) String() string {
	return fmt.Sprintf("%d:%s", u.ID, u.Username) // Custom format
}
