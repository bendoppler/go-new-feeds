package handler

import (
	"context"
	"fmt"
	"news-feed/internal/api/generated/news-feed/userpb"
	"news-feed/internal/service"
	"news-feed/pkg/logger"
)

type GRPCUserHandlerInterface interface {
	Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.LoginResponse, error)
}

// GRPCUserHandler handles requests related to users.
type GRPCUserHandler struct {
	userService service.UserServiceInterface
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
	token, err := h.userService.Login(req.UserName, req.Password)
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
