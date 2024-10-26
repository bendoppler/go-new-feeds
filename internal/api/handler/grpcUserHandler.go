package handler

import (
	"context"
	"fmt"
	"news-feed/internal/api/generated/news-feed/userpb"
	"news-feed/internal/entity"
	"news-feed/internal/service"
	"news-feed/pkg/logger"
	"time"
)

// GRPCUserHandler handles requests related to users.
type GRPCUserHandler struct {
	userpb.UnimplementedUserServiceServer // Embed the unimplemented server
	UserService                           service.UserServiceInterface
}

func (h *GRPCUserHandler) Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	// Check if the context has been canceled
	select {
	case <-ctx.Done():
		return nil, ctx.Err() // Return context error if canceled
	default:
		// Continue processing
	}

	// Simulate user authentication, usually would call a service or database here.
	if req.UserName == "" || req.Password == "" {
		return &userpb.LoginResponse{
			Error: "Username and password are required",
		}, nil
	}

	// Simulate token generation or authentication process
	token, err := h.UserService.Login(req.UserName, req.Password)
	if err != nil {
		logger.LogError(fmt.Sprintf("Login failed: %v", err))
		return &userpb.LoginResponse{
			Error: "Login failed",
		}, nil
	}

	// Return successful gRPC response
	return &userpb.LoginResponse{
		JwtToken: token,
		Error:    "",
	}, nil
}

func (h *GRPCUserHandler) Signup(ctx context.Context, req *userpb.SignupRequest) (*userpb.SignupResponse, error) {
	// Convert the birthday from string to time.Time
	birthday, err := h.convertStringToDate(req.Birthday)
	if err != nil {
		return &userpb.SignupResponse{
			Error: "Invalid birthday format",
		}, nil
	}

	// Convert the gRPC SignupRequest to the entity.User model
	newUser := entity.User{
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Birthday:  birthday,
		Password:  req.Password,
	}

	// Call the user service's Signup method
	token, err := h.UserService.Signup(newUser)
	if err != nil {
		logger.LogError(fmt.Sprintf("Signup failed: %v", err))
		return &userpb.SignupResponse{
			Error: "Signup failed",
		}, nil
	}

	// Return successful gRPC response with the token
	return &userpb.SignupResponse{
		Token: token,
		Error: "",
	}, nil
}

func (h *GRPCUserHandler) convertStringToDate(dateStr string) (time.Time, error) {
	const layout = "2006-01-02" // Define the date layout format
	date, err := time.Parse(layout, dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format: %v", err)
	}
	return date, nil
}

func (h *GRPCUserHandler) EditProfile(ctx context.Context, req *userpb.EditProfileRequest) (*userpb.EditProfileResponse, error) {
	// Validate input (you can also perform input validation in the client or using interceptors)
	if req.FirstName == "" || req.LastName == "" || req.Birthday == "" || req.Password == "" {
		return &userpb.EditProfileResponse{
			Error: "All fields are required",
		}, nil
	}

	// Convert birthday from string to date (assuming you have a utility for this)
	birthday, err := h.convertStringToDate(req.Birthday)
	if err != nil {
		return &userpb.EditProfileResponse{
			Error: "Invalid birthday format",
		}, nil
	}

	// Convert request model to entity model
	user := entity.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Birthday:  birthday,
		Password:  req.Password,
	}

	// Call service to update the profile
	err = h.UserService.EditProfile(user)
	if err != nil {
		return &userpb.EditProfileResponse{
			Error: "Failed to update profile",
		}, nil
	}

	// Return successful response
	return &userpb.EditProfileResponse{
		Message: "Profile updated successfully",
		Error:   "",
	}, nil
}
