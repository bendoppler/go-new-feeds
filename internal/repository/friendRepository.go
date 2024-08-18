package repository

type FriendRepositoryInterface interface {
	GetFriendList(userID string) ([]string, error)
	FollowUser(followerID, followeeID string) error
	UnfollowUser(followerID, followeeID string) error
	GetUserPosts(userID string) ([]string, error)
}
