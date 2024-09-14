package model

// User represents a user in the news feed system.
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// LoginRequest represents the payload for login requests.
type LoginRequest struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

// SignupRequest represents the payload for signup requests.
type SignupRequest struct {
	UserName  string `json:"user_name"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Birthday  string `json:"birthday"`
	Password  string `json:"password"`
}

// ProfileUpdateRequest represents the payload for profile update requests.
type ProfileUpdateRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Birthday  string `json:"birthday"`
	Password  string `json:"password"`
}
