package handler

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	_ "news-feed/docs"
	"news-feed/internal/api/model"
	"news-feed/internal/entity"
	"news-feed/internal/service"
	"news-feed/pkg/logger"
	"news-feed/pkg/middleware"
	"strconv"
	"strings"
	"time"
)

type PostHandlerInterface interface {
	CreatePost() http.HandlerFunc
	GetPost() http.HandlerFunc
	EditPost() http.HandlerFunc
	DeletePost() http.HandlerFunc
	CommentOnPost() http.HandlerFunc
	LikePost() http.HandlerFunc
	PostHandler(w http.ResponseWriter, r *http.Request)
	GetComments() http.HandlerFunc
	GetLikes() http.HandlerFunc
	GetLikesCount() http.HandlerFunc
}

type PostHandler struct {
	postService service.PostServiceInterface
}

func (h *PostHandler) PostHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")

	if len(parts) < 3 || parts[2] != "posts" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if len(parts) == 4 {
			middleware.JWTAuthMiddleware(h.GetPost()).ServeHTTP(w, r)
		} else if len(parts) == 5 {
			if parts[4] == "comments" {
				middleware.JWTAuthMiddleware(h.GetComments()).ServeHTTP(w, r)
			} else if parts[4] == "likes" {
				middleware.JWTAuthMiddleware(h.GetLikes()).ServeHTTP(w, r)
			} else {
				http.NotFound(w, r)
			}
		} else if len(parts) == 6 {
			if parts[4] == "likes" && parts[5] == "count" {
				middleware.JWTAuthMiddleware(h.GetLikesCount()).ServeHTTP(w, r)
			}
		} else {
			http.NotFound(w, r)
		}

	case http.MethodPost:
		if len(parts) == 5 && parts[4] == "comments" {
			middleware.JWTAuthMiddleware(h.CommentOnPost()).ServeHTTP(w, r)
		} else if len(parts) == 5 && parts[4] == "likes" {
			middleware.JWTAuthMiddleware(h.LikePost()).ServeHTTP(w, r)
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}

	case http.MethodPut:
		if len(parts) == 3 {
			middleware.JWTAuthMiddleware(h.EditPost()).ServeHTTP(w, r)
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}

	case http.MethodDelete:
		if len(parts) == 3 {
			middleware.JWTAuthMiddleware(h.DeletePost()).ServeHTTP(w, r)
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// CreatePost creates a new post.
//
// @Summary Create a new post
// @Description Creates a new post with the provided details.
// @Tags posts
// @Accept json
// @Produce json
// @Param request body model.CreatePostRequest true "Post data"
// @Success 200 {object} map[string]string "success response"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/posts [post]
func (h *PostHandler) CreatePost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the JSON request body
		var request model.CreatePostRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			logger.LogError(fmt.Sprintf("Failed to decode JSON: %v", err))
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Retrieve the user ID from the context
		userID, ok := r.Context().Value("userID").(int)
		if !ok {
			logger.LogError(fmt.Sprintf("User ID not found int context"))
			http.Error(w, "User ID not found in context", http.StatusInternalServerError)
			return
		}

		var imageFileName string
		if request.HasImage {
			imageFileName = h.generateUniqueFileName()
		} else {
			imageFileName = ""
		}

		// Call the CreatePost service method
		createdPost, err := h.postService.CreatePost(request.Text, imageFileName, userID)

		// Prepare the response
		response := map[string]interface{}{
			"preSignedURL": createdPost.ContentImagePath,
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to encode response: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// generateUniqueFileName generates a unique file name based on a UUID and the desired file extension.
func (h *PostHandler) generateUniqueFileName() string {
	// Generate a UUID
	uuidString := uuid.New().String()
	// Use a fixed extension or determine it based on some criteria if needed
	extension := ".jpg" // Example extension, modify as needed
	return fmt.Sprintf("%s%s", uuidString, extension)
}

// GetPost retrieves a specific post by its ID.
//
// @Summary Get a specific post
// @Description Retrieves a post by its ID.
// @Tags posts
// @Produce json
// @Param post_id path int true "Post ID"
// @Success 200 {object} entity.Post "Post data"
// @Failure 400 {object} string "Invalid post ID"
// @Failure 404 {object} string "Post not found"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/posts/{post_id} [get]
func (h *PostHandler) GetPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pathParts := strings.Split(r.URL.Path, "/")
		postID, err := strconv.Atoi(pathParts[3])
		if err != nil {
			logger.LogError(fmt.Sprintf("Invalid post ID: %v", err))
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		post, err := h.postService.GetPost(postID)
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to get post: %v", err))
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(post)
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to encode response: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// EditPost updates an existing post.
//
// @Summary Edit an existing post
// @Description Updates an existing post by its ID.
// @Tags posts
// @Accept json
// @Produce json
// @Param post_id path int true "Post ID"
// @Param request body model.EditPostRequest true "Updated post data"
// @Success 200 {object} map[string]string "success response"
// @Failure 400 {object} string "Invalid post ID or request payload"
// @Failure 404 {object} string "Post not found"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/posts/{post_id} [put]
func (h *PostHandler) EditPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pathParts := strings.Split(r.URL.Path, "/")
		postID, err := strconv.Atoi(pathParts[3])
		if err != nil {
			logger.LogError(fmt.Sprintf("Invalid post ID: %v", err))
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		var request model.EditPostRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			logger.LogError(fmt.Sprintf("Failed to decode JSON: %v", err))
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Create an updated post object
		post := entity.Post{
			ID:               postID,
			ContentText:      request.Text,
			ContentImagePath: "",
		}

		// Call service to update the post
		updatedPost, err := h.postService.EditPost(post)
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to update post: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var response map[string]interface{}

		if request.HasImage {
			response["preSignedURL"] = updatedPost.ContentImagePath
		} else {
			response["preSignedURL"] = ""
		}

		// Respond with success
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to encode response: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// DeletePost removes a post.
//
// @Summary Delete a post
// @Description Deletes a post by its ID.
// @Tags posts
// @Param post_id path int true "Post ID"
// @Success 200 {object} map[string]string "success message"
// @Failure 400 {object} string "Invalid post ID"
// @Failure 404 {object} string "Post not found"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/posts/{post_id} [delete]
func (h *PostHandler) DeletePost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pathParts := strings.Split(r.URL.Path, "/")
		postID, err := strconv.Atoi(pathParts[3])
		if err != nil {
			logger.LogError(fmt.Sprintf("Invalid post ID: %v", err))
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		// Retrieve the user ID from the context
		userID, ok := r.Context().Value("userID").(int)
		if !ok {
			logger.LogError(fmt.Sprintf("User ID not found int context"))
			http.Error(w, "User ID not found in context", http.StatusInternalServerError)
			return
		}

		err = h.postService.DeletePost(postID, userID)
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to delete post: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(map[string]string{"msg": "Post deleted successfully"})
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to encode response: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// CommentOnPost adds a comment to a specific post.
//
// @Summary Comment on a post
// @Description Adds a comment to the specified post.
// @Tags posts
// @Accept json
// @Produce json
// @Param post_id path int true "Post ID"
// @Param request body model.CommentOnPostRequest true "Comment data"
// @Success 200 {object} entity.Comment "Comment data"
// @Failure 400 {object} string "Invalid post ID or request payload"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/posts/{post_id}/comments [post]
func (h *PostHandler) CommentOnPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the current user ID from the request context (assumes middleware has set it)
		currentUserID, ok := r.Context().Value("userID").(int)
		if !ok {
			logger.LogError(fmt.Sprintf("Unable to get user id from context"))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		pathParts := strings.Split(r.URL.Path, "/")
		postID, err := strconv.Atoi(pathParts[3])
		if err != nil {
			logger.LogError(fmt.Sprintf("Invalid post id %v", err))
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		var commentRequest model.CommentOnPostRequest
		if err := json.NewDecoder(r.Body).Decode(&commentRequest); err != nil {
			logger.LogError(fmt.Sprintf("Failed to decode JSON: %v", err))
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		createdComment, err := h.postService.CommentOnPost(postID, currentUserID, commentRequest.Text)
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to comment on post: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(createdComment)
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to encode response: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// LikePost allows a user to like a specific post.
//
// @Summary Like a post
// @Description Allows a user to like the specified post.
// @Tags posts
// @Param post_id path int true "Post ID"
// @Success 200 {object} map[string]string "success message"
// @Failure 400 {object} string "Invalid post ID"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/posts/{post_id}/likes [post]
func (h *PostHandler) LikePost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the current user ID from the request context (assumes middleware has set it)
		currentUserID, ok := r.Context().Value("userID").(int)
		if !ok {
			logger.LogError(fmt.Sprintf("Unable to get user id from context"))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		pathParts := strings.Split(r.URL.Path, "/")
		postID, err := strconv.Atoi(pathParts[3])
		if err != nil {
			logger.LogError(fmt.Sprintf("Invalid post id %v", err))
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		err = h.postService.LikePost(postID, currentUserID)
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to like post: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(map[string]string{"msg": "Post liked successfully"})
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to encode response: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// GetComments retrieves comments for a specific post.
//
// @Summary Get comments for a post
// @Description Retrieves comments for the specified post with pagination.
// @Tags posts
// @Produce json
// @Param post_id path int true "Post ID"
// @Param cursor query int false "Cursor for pagination"
// @Param limit query int false "Limit for pagination"
// @Success 200 {array} entity.Comment "List of comments"
// @Failure 400 {object} string "Invalid post ID"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/posts/{post_id}/comments [get]
func (h *PostHandler) GetComments() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pathParts := strings.Split(r.URL.Path, "/")
		postID, err := strconv.Atoi(pathParts[3])
		if err != nil {
			logger.LogError(fmt.Sprintf("Invalid post ID: %v", err))
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
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

		comments, nextCursor, err := h.postService.GetComments(postID, cursor, limit)

		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to get comments: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := struct {
			Comments   []entity.Comment `json:"comments"`
			NextCursor int              `json:"next_cursor"`
		}{
			Comments:   comments,
			NextCursor: nextCursor,
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to encode response: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// GetLikes retrieves likes for a specific post.
//
// @Summary Get likes for a post
// @Description Retrieves likes for the specified post.
// @Tags posts
// @Param post_id path int true "Post ID"
// @Success 200 {array} entity.Like "List of likes"
// @Failure 400 {object} string "Invalid post ID"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/posts/{post_id}/likes [get]
func (h *PostHandler) GetLikes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pathParts := strings.Split(r.URL.Path, "/")
		postID, err := strconv.Atoi(pathParts[3])
		if err != nil {
			logger.LogError(fmt.Sprintf("Invalid post ID: %v", err))
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		limit := 10
		if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil {
			limit = l
		}

		var cursor time.Time
		defaultCursor := time.Unix(0, 0) // Set default cursor to the Unix epoch (1970-01-01)

		// Get cursor from the query parameters
		cursorStr := r.URL.Query().Get("cursor")
		if cursorStr != "" {
			// Attempt to parse the cursor
			parsedCursor, err := time.Parse(time.RFC3339, cursorStr) // Assuming cursor is in RFC3339 format
			if err != nil {
				logger.LogError(fmt.Sprintf("Invalid cursor format: %v", err))
				http.Error(w, "Invalid cursor", http.StatusBadRequest)
				return
			}
			cursor = parsedCursor
		} else {
			// If cursor doesn't exist, set it to the default value
			cursor = defaultCursor
		}
		users, nextCursor, err := h.postService.GetLikes(postID, cursor, limit)
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to get likes for post: %v", err))
			return
		}
		// Prepare the response including the nextCursor
		response := struct {
			Users      []entity.User `json:"users"`
			NextCursor *time.Time    `json:"next_cursor"`
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

// GetLikesCount retrieves the count of likes for a specific post.
//
// @Summary Get likes count for a post
// @Description Retrieves the total count of likes for the specified post.
// @Tags posts
// @Param post_id path int true "Post ID"
// @Success 200 {object} map[string]int "Count of likes"
// @Failure 400 {object} string "Invalid post ID"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/posts/{post_id}/likes/count [get]
func (h *PostHandler) GetLikesCount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pathParts := strings.Split(r.URL.Path, "/")
		postID, err := strconv.Atoi(pathParts[3])
		if err != nil {
			logger.LogError(fmt.Sprintf("Invalid post ID: %v", err))
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		likeCount, err := h.postService.GetLikeCount(postID)
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to get likes for post: %v", err))
			http.Error(w, "Failed to retrieve like count", http.StatusInternalServerError)
			return
		}

		// Respond with the like count
		response := map[string]int{"like_count": likeCount}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed encode response: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
