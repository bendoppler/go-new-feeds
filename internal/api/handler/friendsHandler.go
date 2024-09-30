package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"news-feed/internal/entity"
	"news-feed/pkg/logger"
	"news-feed/pkg/middleware"
	"strconv"
	"strings"

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
	friendsService service.FriendsServiceInterface
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
// API: /v1/friends/{user_id}?cursor=12345&limit=10
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

		//logger.LogInfo(fmt.Sprintf("Getting friends for user %d and cursor: %d", userID, cursor))

		users, nextCursor, err := h.friendsService.GetFriends(userID, limit, cursor)
		if err != nil {
			logger.LogError(fmt.Sprintf("Get followers failed %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//logger.LogInfo(fmt.Sprintf("User count: %v and next cursor: %v", len(users), nextCursor))

		// Prepare the response including the nextCursor
		response := struct {
			Users      []entity.User `json:"users"`
			NextCursor int           `json:"next_cursor"`
		}{
			Users:      users,
			NextCursor: nextCursor,
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

// FollowUser handles POST requests for following a user.
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

		// Call the service method to follow the target user
		msg, err := h.friendsService.FollowUser(currentUserID, targetUserID)
		if err != nil {
			logger.LogError(fmt.Sprintf("Follow user failed %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(map[string]string{"msg": msg})
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed encode response: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// UnfollowUser handles DELETE requests for unfollowing a user.
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

		// Call the service method to unfollow the target user
		msg, err := h.friendsService.UnfollowUser(currentUserID, targetUserID)
		if err != nil {
			logger.LogError(fmt.Sprintf("Unfollow user failed %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(map[string]string{"msg": msg})
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed encode response: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// GetUserPosts handles GET requests for retrieving posts by a user.
// API: /v1/friends/{user_id}/posts?cursor=12345&limit=10
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

		posts, nextCursor, err := h.friendsService.GetUserPosts(userID, limit, cursor)
		if err != nil {
			logger.LogError(fmt.Sprintf("Get user posts failed %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Prepare the response including the nextCursor
		response := struct {
			Posts      []entity.Post `json:"posts"`
			NextCursor int           `json:"next_cursor"`
		}{
			Posts:      posts,
			NextCursor: nextCursor,
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
