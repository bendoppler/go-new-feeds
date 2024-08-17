package entity

import "time"

// User represents a user in the system, used internally for business logic and database operations.
type User struct {
	ID        int
	UserName  string
	Email     string
	FirstName string
	LastName  string
	Birthday  time.Time
	Password  string
}
