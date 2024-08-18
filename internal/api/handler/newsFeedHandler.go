package handler

import (
	"encoding/json"
	"net/http"

	"news-feed/internal/service"
)

type NewsFeedHandlerInterface interface {
	GetNewsfeed(w http.ResponseWriter, r *http.Request)
}

type NewsfeedHandler struct {
	newsFeedService service.NewsFeedServiceInterface
}

// GetNewsfeed handles GET requests for retrieving newsfeed posts.
func (h *NewsfeedHandler) GetNewsfeed(w http.ResponseWriter, r *http.Request) {
	posts, err := h.newsFeedService.GetNewsfeedPosts()
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
