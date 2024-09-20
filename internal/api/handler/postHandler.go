package handler

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"news-feed/internal/api/model"
	"news-feed/internal/entity"
	"news-feed/internal/service"
	"news-feed/pkg/logger"
	"news-feed/pkg/middleware"
	"strconv"
	"strings"
)

type PostHandlerInterface interface {
	CreatePost() http.HandlerFunc
	GetPost() http.HandlerFunc
	EditPost() http.HandlerFunc
	DeletePost() http.HandlerFunc
	CommentOnPost() http.HandlerFunc
	LikePost() http.HandlerFunc
	PostHandler(w http.ResponseWriter, r *http.Request)
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
		if len(parts) == 3 {
			h.GetPost()
		} else {
			http.NotFound(w, r)
		}

	case http.MethodPost:
		if len(parts) == 3 {
			middleware.JWTAuthMiddleware(h.CreatePost()).ServeHTTP(w, r)
		} else if len(parts) == 4 && parts[3] == "comments" {
			middleware.JWTAuthMiddleware(h.CommentOnPost()).ServeHTTP(w, r)
		} else if len(parts) == 4 && parts[3] == "likes" {
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
		preSignedURL, isSuccess, err := h.postService.CreatePost(request.Text, imageFileName, userID)

		// Prepare the response
		response := map[string]interface{}{
			"preSignedURL": preSignedURL,
			"is_success":   isSuccess,
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

func (h *PostHandler) GetPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID, err := strconv.Atoi(mux.Vars(r)["postId"])
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		post, err := h.postService.GetPost(postID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(post)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (h *PostHandler) EditPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID, err := strconv.Atoi(mux.Vars(r)["postId"])
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		// Parse the multipart form
		err = r.ParseMultipartForm(10 << 20) // 10MB limit
		if err != nil {
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		// Extract text and file from form
		text := r.FormValue("text")
		imageFile, _, err := r.FormFile("image")
		if err != nil && err.Error() != "http: no such file" {
			http.Error(w, "Unable to get image file", http.StatusBadRequest)
			return
		}

		// Prepare image file information
		var imageFileName string
		if imageFile != nil {
			imageFileName = "post_" + strconv.Itoa(postID) // Generate a unique name for the image
			imageFileURL, uploadErr := h.postService.UploadImage(imageFileName, imageFile)
			if uploadErr != nil {
				http.Error(w, "Failed to upload image", http.StatusInternalServerError)
				return
			}
			imageFileName = imageFileURL
		}

		// Create an updated post object
		post := entity.Post{
			ID:               postID,
			ContentText:      text,
			ContentImagePath: imageFileName,
		}

		// Call service to update the post
		err = h.postService.EditPost(post)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Respond with success
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(map[string]string{"msg": "Post updated successfully"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (h *PostHandler) DeletePost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID, err := strconv.Atoi(mux.Vars(r)["postId"])
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		err = h.postService.DeletePost(postID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(map[string]string{"msg": "Post deleted successfully"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (h *PostHandler) CommentOnPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID, err := strconv.Atoi(mux.Vars(r)["postId"])
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		var commentRequest model.CommentOnPostRequest
		if err := json.NewDecoder(r.Body).Decode(&commentRequest); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		err = h.postService.CommentOnPost(postID, commentRequest.Text)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(map[string]string{"msg": "Comment added successfully"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (h *PostHandler) LikePost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID, err := strconv.Atoi(mux.Vars(r)["postId"])
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		err = h.postService.LikePost(postID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(map[string]string{"msg": "Post liked successfully"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
