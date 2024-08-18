package model

type FollowRequest struct {
	FollowerID string `json:"followerId"`
	FolloweeID string `json:"followeeId"`
}

type FriendListResponse struct {
	UserID    string   `json:"userId"`
	FriendIDs []string `json:"friendIds"`
}

type UserPostsResponse struct {
	UserID string   `json:"userId"`
	Posts  []string `json:"posts"` // This can be more detailed based on your post structure.
}
