package model

// CreatePostRequest represents the request payload for creating a new post.
//
// @Description Request payload for creating a new post in the news feed.
// @Model
type CreatePostRequest struct {
	// Text is the content of the post.
	// @example "This is a new post"
	Text string `json:"text"` // The text content of the post

	// HasImage indicates whether the post includes an image.
	// @example true
	HasImage bool `json:"hasImage"` // Flag to indicate if the post contains an image
}

// EditPostRequest represents the request payload for editing an existing post.
type EditPostRequest struct {
	Text     string `json:"text"`
	HasImage bool   `json:"hasImage"`
}

// DeletePostRequest represents the request payload for deleting a post.
type DeletePostRequest struct {
	PostID int `json:"post_id"`
}

// CommentOnPostRequest represents the request payload for commenting on a post.
type CommentOnPostRequest struct {
	Text string `json:"text"`
}

// LikePostRequest represents the request payload for liking a post.
type LikePostRequest struct{}

// GetPostResponse represents the response payload for getting a post.
type GetPostResponse struct {
	ID       int      `json:"id"`
	Text     string   `json:"text"`
	Image    string   `json:"image"` // URL or path to the image
	Comments []string `json:"comments"`
	Likes    int      `json:"likes"`
}
