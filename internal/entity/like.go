package entity

import "time"

type Like struct {
	PostID    int
	UserID    int
	CreatedAt time.Time
}
