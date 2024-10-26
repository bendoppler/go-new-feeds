package handler

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"news-feed/internal/api/generated/news-feed/friendspb"
	"news-feed/internal/service"
	"news-feed/pkg/logger"
)

type GRPCFriendsHandler struct {
	friendspb.UnimplementedFriendsServiceServer // Embed the unimplemented server
	FriendsService                              service.FriendsServiceInterface
}

func (h *GRPCFriendsHandler) GetFriends(ctx context.Context, req *friendspb.GetFriendsRequest) (*friendspb.GetFriendsResponse, error) {
	userID := req.GetUserId()
	limit := req.GetLimit()
	cursor := req.GetCursor()

	// Fetch friends from the service
	users, nextCursor, err := h.FriendsService.GetFriends(int(userID), int(limit), int(cursor))
	if err != nil {
		return nil, err // Handle the error appropriately
	}

	// Prepare the response
	response := &friendspb.GetFriendsResponse{
		Users:      make([]*friendspb.User, len(users)),
		NextCursor: int32(nextCursor),
	}

	// Map entity.User to friendspb.User
	for i, user := range users {
		response.Users[i] = &friendspb.User{
			Id:        int32(user.ID),
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Username:  user.Username,
		}
	}

	return response, nil
}

func (h *GRPCFriendsHandler) FollowUser(ctx context.Context, req *friendspb.FollowUserRequest) (*friendspb.FollowUserResponse, error) {
	currentUserID := req.GetCurrentUserId()
	targetUserID := req.GetTargetUserId()

	// Call the service method to follow the target user
	msg, err := h.FriendsService.FollowUser(int(currentUserID), int(targetUserID))
	if err != nil {
		logger.LogError(fmt.Sprintf("Follow user failed %v", err))
		return nil, err // Return the error to gRPC
	}

	// Prepare the response
	response := &friendspb.FollowUserResponse{
		Msg: msg,
	}

	return response, nil
}

func (h *GRPCFriendsHandler) UnfollowUser(ctx context.Context, req *friendspb.UnfollowUserRequest) (*friendspb.UnfollowUserResponse, error) {
	currentUserID := req.GetCurrentUserId()
	targetUserID := req.GetTargetUserId()

	// Call the service method to unfollow the target user
	msg, err := h.FriendsService.UnfollowUser(int(currentUserID), int(targetUserID))
	if err != nil {
		logger.LogError(fmt.Sprintf("Unfollow user failed %v", err))
		return nil, err // Return the error to gRPC
	}

	// Prepare the response
	response := &friendspb.UnfollowUserResponse{
		Msg: msg,
	}

	return response, nil
}

func (h *GRPCFriendsHandler) GetUserPosts(ctx context.Context, req *friendspb.GetUserPostsRequest) (*friendspb.GetUserPostsResponse, error) {
	userID := req.GetUserId()
	limit := req.GetLimit()
	cursor := req.GetCursor()

	// Call the service method to get user posts
	posts, nextCursor, err := h.FriendsService.GetUserPosts(int(userID), int(limit), int(cursor))
	if err != nil {
		logger.LogError(fmt.Sprintf("Get user posts failed %v", err))
		return nil, err // Return the error to gRPC
	}

	// Convert posts to gRPC format
	grpcPosts := make([]*friendspb.Post, len(posts))
	for i, post := range posts {
		grpcPosts[i] = &friendspb.Post{
			Id:               int32(post.ID),                  // assuming Post has an ID field
			UserId:           int32(post.UserID),              // Map UserID
			ContentText:      post.ContentText,                // Map ContentText
			ContentImagePath: post.ContentImagePath,           // Map ContentImagePath
			CreatedAt:        timestamppb.New(post.CreatedAt), // Convert time.Time to protobuf Timestamp
		}
	}

	// Prepare the response
	response := &friendspb.GetUserPostsResponse{
		Posts:      grpcPosts,
		NextCursor: int32(nextCursor),
	}

	return response, nil
}
