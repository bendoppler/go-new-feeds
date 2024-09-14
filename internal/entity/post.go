package entity

import "time"

type Post struct {
	ID               int
	UserID           int
	ContentText      string
	ContentImagePath string
	CreatedAt        time.Time
}
