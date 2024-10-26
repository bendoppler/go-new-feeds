package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"news-feed/internal/api/generated/news-feed/friendspb"
	"news-feed/pkg/logger"
	"news-feed/pkg/middleware"
	"strconv"
	"strings"

	_ "news-feed/docs"
	"news-feed/internal/service"
)

type FriendsHandlerInterface interface {
	GetFriends() http.HandlerFunc
	FollowUser() http.HandlerFunc
	UnfollowUser() http.HandlerFunc
	GetUserPosts() http.HandlerFunc
	FriendsHandler(w http.ResponseWriter, r *http.Request)
}

type FriendsHandler struct {
	friendsService     service.FriendsServiceInterface
	grpcFriendsHandler friendspb.FriendsServiceServer
}

// FriendsHandler handles all requests under /v1/friends/
func (h *FriendsHandler) FriendsHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")

	if len(parts) < 3 || parts[2] != "friends" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if len(parts) == 4 {
			middleware.JWTAuthMiddleware(h.GetFriends()).ServeHTTP(w, r)
		} else if len(parts) == 5 && parts[4] == "posts" {
			middleware.JWTAuthMiddleware(h.GetUserPosts()).ServeHTTP(w, r)
		} else {
			http.NotFound(w, r)
		}

	case http.MethodPost:
		if len(parts) == 4 {
			middleware.JWTAuthMiddleware(h.FollowUser()).ServeHTTP(w, r)
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}

	case http.MethodDelete:
		if len(parts) == 4 {
			middleware.JWTAuthMiddleware(h.UnfollowUser()).ServeHTTP(w, r)
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GetFriends handles GET requests for retrieving a list of friends.
// @Summary Get a list of friends for a user
// @Description Get friends of the user by user ID
// @Tags friends
// @Accept  json
// @Produce  json
// @Param   user_id  path     int true  "User ID"
// @Param   cursor   query    int false "Cursor for pagination"
// @Param   limit    query    int false "Limit of results"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/friends/{user_id} [get]
func (h *FriendsHandler) GetFriends() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pathParts := strings.Split(r.URL.Path, "/")
		userID, err := strconv.Atoi(pathParts[3])
		if err != nil {
			logger.LogError(fmt.Sprintf("Invalid user id %v", err))
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		limit := 10
		if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil {
			limit = l
		}

		cursor := 0
		if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
			cursor, err = strconv.Atoi(cursorStr)
			if err != nil {
				logger.LogError(fmt.Sprintf("Invalid cursor %v", err))
				http.Error(w, "Invalid cursor", http.StatusBadRequest)
				return
			}
		}

		req := friendspb.GetFriendsRequest{
			UserId: int32(userID),
			Limit:  int32(limit),
			Cursor: int32(cursor),
		}

		response, err := h.grpcFriendsHandler.GetFriends(context.Background(), &req)
		if err != nil {
			logger.LogError(fmt.Sprintf("Get followers failed %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed encode response: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (h *FriendsHandler) FollowUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the current user ID from the request context (assumes middleware has set it)
		currentUserID, ok := r.Context().Value("userID").(int)
		if !ok {
			logger.LogError(fmt.Sprintf("Unable to get user id from context"))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse the target user ID from the URL parameters
		pathParts := strings.Split(r.URL.Path, "/")
		targetUserID, err := strconv.Atoi(pathParts[3])
		if err != nil {
			logger.LogError(fmt.Sprintf("Invalid user id %v", err))
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		req := friendspb.FollowUserRequest{
			CurrentUserId: int32(currentUserID),
			TargetUserId:  int32(targetUserID),
		}

		// Call the service method to follow the target user
		response, err := h.grpcFriendsHandler.FollowUser(context.Background(), &req)
		if err != nil {
			logger.LogError(fmt.Sprintf("Follow user failed %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed encode response: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// UnfollowUser handles DELETE requests for unfollowing a user.
// @Summary Unfollow a user
// @Description Unfollow a user by providing the target user ID
// @Tags friends
// @Accept  json
// @Produce  json
// @Param   user_id  path     int true  "Target User ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/friends/{user_id} [delete]
func (h *FriendsHandler) UnfollowUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the current user ID from the request context (assumes middleware has set it)
		currentUserID, ok := r.Context().Value("userID").(int)
		if !ok {
			logger.LogError(fmt.Sprintf("Unable to get user id from context"))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse the target user ID from the URL parameters
		pathParts := strings.Split(r.URL.Path, "/")
		targetUserID, err := strconv.Atoi(pathParts[3])
		if err != nil {
			logger.LogError(fmt.Sprintf("Invalid user id %v", err))
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		req := friendspb.UnfollowUserRequest{
			CurrentUserId: int32(currentUserID),
			TargetUserId:  int32(targetUserID),
		}

		// Call the service method to unfollow the target user
		response, err := h.grpcFriendsHandler.UnfollowUser(context.Background(), &req)
		if err != nil {
			logger.LogError(fmt.Sprintf("Unfollow user failed %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed encode response: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// GetUserPosts handles GET requests for retrieving posts by a user.
// @Summary Get posts for a user's friends
// @Description Get posts made by a user's friends, with pagination
// @Tags friends
// @Accept  json
// @Produce  json
// @Param   user_id  path     int true  "User ID"
// @Param   cursor   query    int false "Cursor for pagination"
// @Param   limit    query    int false "Limit of results"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/friends/{user_id}/posts [get]
func (h *FriendsHandler) GetUserPosts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pathParts := strings.Split(r.URL.Path, "/")
		userID, err := strconv.Atoi(pathParts[3])
		if err != nil {
			logger.LogError(fmt.Sprintf("Invalid user id %v", err))
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		limit := 10
		if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil {
			limit = l
		}

		cursor := 0
		if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
			cursor, err = strconv.Atoi(cursorStr)
			if err != nil {
				logger.LogError(fmt.Sprintf("Invalid cursor %v", err))
				http.Error(w, "Invalid cursor", http.StatusBadRequest)
				return
			}
		}

		req := friendspb.GetUserPostsRequest{
			UserId: int32(userID),
			Limit:  int32(limit),
			Cursor: int32(cursor),
		}
		response, err := h.grpcFriendsHandler.GetUserPosts(context.Background(), &req)
		if err != nil {
			logger.LogError(fmt.Sprintf("Get user posts failed %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed encode response: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
