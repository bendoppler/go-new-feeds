package entity

import "time"

// Post represents a post in the news feed.
//
// @Description Represents a post created by a user in the news feed.
// @Model
type Post struct {
	ID               int       `json:"id"`
	UserID           int       `json:"user_id"`
	ContentText      string    `json:"content_text"`
	ContentImagePath string    `json:"content_image_path"`
	CreatedAt        time.Time `json:"created_at"`
}
