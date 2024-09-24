package handler

import (
	"encoding/json"
	"net/http"
	"news-feed/pkg/middleware"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
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

	if len(parts) < 3 || parts[1] != "friends" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if len(parts) == 3 {
			middleware.JWTAuthMiddleware(h.GetFriends()).ServeHTTP(w, r)
		} else if len(parts) == 4 && parts[3] == "posts" {
			middleware.JWTAuthMiddleware(h.GetUserPosts()).ServeHTTP(w, r)
		} else {
			http.NotFound(w, r)
		}

	case http.MethodPost:
		if len(parts) == 3 {
			middleware.JWTAuthMiddleware(h.FollowUser()).ServeHTTP(w, r)
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}

	case http.MethodDelete:
		if len(parts) == 3 {
			middleware.JWTAuthMiddleware(h.UnfollowUser()).ServeHTTP(w, r)
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GetFriends handles GET requests for retrieving a list of friends.
func (h *FriendsHandler) GetFriends() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.Atoi(mux.Vars(r)["userId"])
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		users, err := h.friendsService.GetFriends(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(users)
		if err != nil {
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
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse the target user ID from the URL parameters
		targetUserID, err := strconv.Atoi(mux.Vars(r)["userId"])
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		// Call the service method to follow the target user
		msg, err := h.friendsService.FollowUser(currentUserID, targetUserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(map[string]string{"msg": msg})
		if err != nil {
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
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse the target user ID from the URL parameters
		targetUserID, err := strconv.Atoi(mux.Vars(r)["userId"])
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		// Call the service method to unfollow the target user
		msg, err := h.friendsService.UnfollowUser(currentUserID, targetUserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(map[string]string{"msg": msg})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// GetUserPosts handles GET requests for retrieving posts by a user.
func (h *FriendsHandler) GetUserPosts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.Atoi(mux.Vars(r)["userId"])
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		posts, err := h.friendsService.GetUserPosts(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(posts)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
