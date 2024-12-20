package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"news-feed/internal/api/generated/news-feed/newsfeedpb"

	_ "news-feed/docs"
	_ "news-feed/internal/entity"
)

type NewsFeedHandlerInterface interface {
	GetNewsfeed() http.HandlerFunc
}

type NewsfeedHandler struct {
	newsFeedService newsfeedpb.NewsfeedServiceClient
}

// GetNewsfeed handles GET requests for retrieving newsfeed posts.
// @Summary Get news feed
// @Description Get the latest posts from user's friends.
// @Tags NewsFeed
// @Produce json
// @Success 200 {array} entity.Post "List of posts"
// @Failure 500 {object} error "Internal server error"
// @Router /v1/newsfeed [get]
func (h *NewsfeedHandler) GetNewsfeed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		posts, err := h.newsFeedService.GetNewsfeed(context.Background(), &newsfeedpb.GetNewsfeedRequest{})
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
