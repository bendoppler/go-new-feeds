package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"mime/multipart"
	"net/http"
	"news-feed/internal/api/model"
	"news-feed/internal/entity"
	"news-feed/internal/service"
	"strconv"
)

type PostHandlerInterface interface {
	CreatePost(w http.ResponseWriter, r *http.Request)
	GetPost(w http.ResponseWriter, r *http.Request)
	EditPost(w http.ResponseWriter, r *http.Request)
	DeletePost(w http.ResponseWriter, r *http.Request)
	CommentOnPost(w http.ResponseWriter, r *http.Request)
	LikePost(w http.ResponseWriter, r *http.Request)
}

type PostHandler struct {
	postService service.PostService
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	err := r.ParseMultipartForm(10 << 20) // Limit to 10 MB
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Retrieve text and image file from the form
	text := r.FormValue("text")
	imageFile, _, err := r.FormFile("image")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		http.Error(w, "Failed to retrieve file", http.StatusBadRequest)
		return
	}
	defer func(imageFile multipart.File) {
		err := imageFile.Close()
		if err != nil {
			fmt.Printf("Failed to close image: %f\n", err)
			return
		}
	}(imageFile)

	// Create a unique filename for the image if provided
	imageFileName := h.generateUniqueFileName()

	// Call the CreatePost service method
	msg, isSuccess, errCode := h.postService.CreatePost(text, imageFileName, imageFile)

	// Prepare the response
	response := map[string]interface{}{
		"msg":        msg,
		"is_success": isSuccess,
		"err_code":   errCode,
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
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

func (h *PostHandler) EditPost(w http.ResponseWriter, r *http.Request) {
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

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
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

func (h *PostHandler) CommentOnPost(w http.ResponseWriter, r *http.Request) {
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

func (h *PostHandler) LikePost(w http.ResponseWriter, r *http.Request) {
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
