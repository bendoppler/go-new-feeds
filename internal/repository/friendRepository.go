package repository

import (
	"database/sql"
	"fmt"
	"news-feed/internal/entity"
)

type FriendsRepositoryInterface interface {
	GetFriends(userID int, limit int, cursor int) ([]entity.User, int, error)
	FollowUser(currentUserID int, followedUserID int) error
	UnfollowUser(currentUserID int, unfollowedUserID int) error
}

type FriendsRepository struct {
	db *sql.DB
}

// GetFriends retrieves the list of friends for a user.
func (r *FriendsRepository) GetFriends(userID int, limit int, cursor int) ([]entity.User, int, error) {
	// Query to get followers with pagination using cursor
	rows, err := r.db.Query(
		"SELECT u.id, u.first_name, u.last_name, u.email, u.user_name, uu.fk_follower_id FROM user u "+
			"JOIN user_user uu ON u.id = uu.fk_follower_id "+
			"WHERE uu.fk_user_id = ? AND uu.fk_follower_id > ? ORDER BY uu.fk_follower_id ASC LIMIT ?",
		userID, cursor, limit,
	)
	if err != nil {
		return nil, 0, err
	}
	defer func(rows *sql.Rows) {
		if err := rows.Close(); err != nil {
			fmt.Printf("Failed to close rows: %v\n", err)
		}
	}(rows)

	var users []entity.User
	var nextCursor int

	// Process rows and set nextCursor
	for rows.Next() {
		var user entity.User
		var followerID int // This will serve as the next cursor

		// Scan all the required fields including fk_follower_id
		if err := rows.Scan(
			&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Username, &followerID,
		); err != nil {
			return nil, 0, err
		}

		users = append(users, user)
		nextCursor = followerID // Update nextCursor with the last follower's ID
	}

	// If no rows were returned, nextCursor remains 0, indicating no more data.
	return users, nextCursor, nil
}

// FollowUser follows a user.
func (r *FriendsRepository) FollowUser(currentUserID int, followedUserID int) error {
	// This example assumes you want to follow user ID 2, modify accordingly
	_, err := r.db.Exec(
		"INSERT INTO user_user (fk_user_id, fk_follower_id) VALUES (?, ?)", currentUserID, followedUserID,
	)
	return err
}

// UnfollowUser unfollows a user.
func (r *FriendsRepository) UnfollowUser(currentUserID int, unfollowedUserID int) error {
	_, err := r.db.Exec(
		"DELETE FROM user_user WHERE fk_user_id = ? AND fk_follower_id = ?", currentUserID, unfollowedUserID,
	)
	return err
}
