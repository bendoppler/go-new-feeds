package entity

type Friend struct {
	ID         string `json:"id"`
	FollowerID string `json:"follower_id"`
	FolloweeID string `json:"followee_id"`
}
