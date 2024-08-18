package handler

import (
	"encoding/json"
	"net/http"
	"news-feed/internal/api/model"
	"news-feed/internal/service"
)

type PostHandlerInterface interface {
	CreatePost(w http.ResponseWriter, r *http.Request)
}

type PostHandler struct {
	postService service.PostService
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var post model.Post
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	msg, success, errCode := h.postService.CreatePost(&post)
	if !success {
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(
		map[string]interface{}{
			"msg":        msg,
			"is_success": success,
			"err_code":   errCode,
		},
	)
}
