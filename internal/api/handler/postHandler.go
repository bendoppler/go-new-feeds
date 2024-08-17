package handler

import (
	"encoding/json"
	"net/http"
	"news-feed/internal/api/model"
	"news-feed/internal/service"
)

// PostHandler handles requests related to posts.
type PostHandler struct {
	postService service.PostService
}

// NewPostHandler creates a new PostHandler.
func NewPostHandler(postService service.PostService) *PostHandler {
	return &PostHandler{
		postService: postService,
	}
}

// GetPosts handles GET requests to fetch all posts.
func (h *PostHandler) GetPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := h.postService.GetAllPosts()
	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

// CreatePost handles POST requests to create a new post.
func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var post model.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := h.postService.CreatePost(post); err != nil {
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
