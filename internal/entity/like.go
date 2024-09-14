package entity

import "time"

type Like struct {
	ID        int
	PostID    int
	UserID    int
	CreatedAt time.Time
}
