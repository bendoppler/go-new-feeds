package handler

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/uuid"
	"log"
	_ "news-feed/docs"
	"news-feed/internal/api/generated/news-feed/postpb"
	"news-feed/internal/entity"
	"news-feed/internal/service"
	"time"
)

type GRPCPostHandler struct {
	postpb.UnimplementedPostServiceServer // Embed the unimplemented server
	PostService                           service.PostServiceInterface
}

func (h *GRPCPostHandler) CreatePost(ctx context.Context, req *postpb.CreatePostRequest) (*postpb.CreatePostResponse, error) {
	// Retrieve the user ID from the context
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		log.Printf("User ID not found in context")
		return nil, fmt.Errorf("User ID not found in context")
	}

	var imageFileName string
	if req.HasImage {
		imageFileName = h.generateUniqueFileName()
	} else {
		imageFileName = ""
	}

	// Call the CreatePost service method
	createdPost, err := h.PostService.CreatePost(req.Text, imageFileName, userID)
	if err != nil {
		log.Printf("Failed to create post: %v", err)
		return nil, fmt.Errorf("failed to create post: %v", err)
	}

	// Prepare the response
	response := &postpb.CreatePostResponse{
		PreSignedURL: createdPost.ContentImagePath,
	}

	return response, nil
}

// generateUniqueFileName generates a unique file name based on a UUID and the desired file extension.
func (h *GRPCPostHandler) generateUniqueFileName() string {
	// Generate a UUID
	uuidString := uuid.New().String()
	// Use a fixed extension or determine it based on some criteria if needed
	extension := ".jpg" // Example extension, modify as needed
	return fmt.Sprintf("%s%s", uuidString, extension)
}

func (h *GRPCPostHandler) GetPost(ctx context.Context, req *postpb.GetPostRequest) (*postpb.GetPostResponse, error) {
	postID := req.PostId

	// Call the GetPost service method
	post, err := h.PostService.GetPost(int(postID))
	if err != nil {
		log.Printf("Failed to get post: %v", err)
		return nil, fmt.Errorf("failed to get post: %v", err)
	}

	// Prepare the response
	response := &postpb.GetPostResponse{
		Id:               int32(post.ID),                      // Convert to int32 for gRPC
		UserId:           int32(post.UserID),                  // Convert to int32 for gRPC
		ContentText:      post.ContentText,                    // Post content text
		ContentImagePath: post.ContentImagePath,               // Image URL or path
		CreatedAt:        post.CreatedAt.Format(time.RFC3339), // Format time.Time to string in RFC3339
	}

	return response, nil
}

func (h *GRPCPostHandler) EditPost(ctx context.Context, req *postpb.EditPostRequest) (*postpb.EditPostResponse, error) {
	postID := req.PostId // postID is of type int32

	// Create an updated post object
	post := entity.Post{
		ID:               int(postID),     // Convert to int if needed
		ContentText:      req.ContentText, // Content text from request
		ContentImagePath: "",              // Placeholder for image path
	}

	// Call service to update the post
	updatedPost, err := h.PostService.EditPost(post)
	if err != nil {
		log.Printf("Failed to update post: %v", err)
		return nil, fmt.Errorf("failed to update post: %v", err)
	}

	// Prepare the response
	response := &postpb.EditPostResponse{
		PreSignedUrl: "", // Initialize with an empty URL
	}

	if req.HasImage {
		response.PreSignedUrl = updatedPost.ContentImagePath // Set the pre-signed URL if the request indicates an image
	}

	return response, nil
}

func (h *GRPCPostHandler) DeletePost(ctx context.Context, req *postpb.DeletePostRequest) (*postpb.DeletePostResponse, error) {
	postID := req.PostId
	userID := req.UserId

	// Call the DeletePost service method
	err := h.PostService.DeletePost(int(postID), int(userID))
	if err != nil {
		log.Printf("Failed to delete post: %v", err)
		return nil, fmt.Errorf("failed to delete post: %v", err)
	}

	// Prepare the response
	response := &postpb.DeletePostResponse{
		Msg: "Post deleted successfully", // Confirmation message
	}

	return response, nil
}

func (h *GRPCPostHandler) CommentOnPost(ctx context.Context, req *postpb.CommentOnPostRequest) (*postpb.CommentOnPostResponse, error) {
	postID := req.PostId
	userID := req.UserId
	commentText := req.Text

	// Call the CommentOnPost service method
	createdComment, err := h.PostService.CommentOnPost(int(postID), int(userID), commentText)
	if err != nil {
		log.Printf("Failed to comment on post: %v", err)
		return nil, fmt.Errorf("failed to comment on post: %v", err)
	}

	// Prepare the response
	response := &postpb.CommentOnPostResponse{
		CommentId: int32(createdComment.ID),                      // Assuming createdComment has an ID field
		Text:      createdComment.Content,                        // Comment text
		UserId:    int32(createdComment.UserID),                  // ID of the user who commented
		CreatedAt: createdComment.CreatedAt.Format(time.RFC3339), // Format timestamp
	}

	return response, nil
}

func (h *GRPCPostHandler) LikePost(ctx context.Context, req *postpb.LikePostRequest) (*postpb.LikePostResponse, error) {
	postID := req.PostId
	userID := req.UserId

	// Call the LikePost service method
	err := h.PostService.LikePost(int(postID), int(userID))
	if err != nil {
		log.Printf("Failed to like post: %v", err)
		return nil, fmt.Errorf("failed to like post: %v", err)
	}

	// Prepare the response
	response := &postpb.LikePostResponse{
		Message: "Post liked successfully",
	}

	return response, nil
}

func (h *GRPCPostHandler) GetComments(ctx context.Context, req *postpb.GetCommentsRequest) (*postpb.GetCommentsResponse, error) {
	postID := req.PostId
	cursor := req.Cursor
	limit := req.Limit

	// Call the GetComments service method
	comments, nextCursor, err := h.PostService.GetComments(int(postID), int(cursor), int(limit))
	if err != nil {
		log.Printf("Failed to get comments: %v", err)
		return nil, fmt.Errorf("failed to get comments: %v", err)
	}

	// Prepare the response
	response := &postpb.GetCommentsResponse{
		NextCursor: int32(nextCursor),
	}

	// Populate the comments
	for _, comment := range comments {
		response.Comments = append(
			response.Comments, &postpb.Comment{
				Id:        int32(comment.ID),
				UserId:    int32(comment.UserID),
				Text:      comment.Content,
				CreatedAt: comment.CreatedAt.Format(time.RFC3339), // Format timestamp
			},
		)
	}

	return response, nil
}

func (h *GRPCPostHandler) GetLikes(ctx context.Context, req *postpb.GetLikesRequest) (*postpb.GetLikesResponse, error) {
	postID := req.PostId
	cursor := req.Cursor
	limit := req.Limit

	// Call the GetLikes service method
	users, nextCursor, err := h.PostService.GetLikes(int(postID), *parseCursor(cursor), int(limit))
	if err != nil {
		log.Printf("Failed to get likes: %v", err)
		return nil, fmt.Errorf("failed to get likes: %v", err)
	}

	// Prepare the response
	response := &postpb.GetLikesResponse{
		NextCursor: nextCursor.Format(time.RFC3339), // Convert *time.Time to google.protobuf.Timestamp
	}

	// Populate the users
	for _, user := range users {
		response.Users = append(
			response.Users, &postpb.User{
				Id:             int32(user.ID),
				HashedPassword: user.HashedPassword,
				Salt:           user.Salt,
				FirstName:      user.FirstName,
				LastName:       user.LastName,
				Birthday:       toTimestamp(&user.Birthday), // Convert time.Time to google.protobuf.Timestamp
				Email:          user.Email,
				Username:       user.Username,
			},
		)
	}

	return response, nil
}

// Convert time.Time to google.protobuf.Timestamp
func toTimestamp(t *time.Time) *timestamp.Timestamp {
	if t == nil || t.IsZero() {
		return nil // or return a default timestamp if needed
	}
	return &timestamp.Timestamp{
		Seconds: t.Unix(),
		Nanos:   int32(t.Nanosecond()),
	}
}

// Parse the cursor string into a time.Time (or handle it according to your logic)
func parseCursor(cursor string) *time.Time {
	if cursor == "" {
		return nil // Return nil if the cursor is empty
	}
	parsedCursor, err := time.Parse(time.RFC3339, cursor)
	if err != nil {
		return nil // Return nil if parsing fails
	}
	return &parsedCursor
}

func (h *GRPCPostHandler) GetLikesCount(ctx context.Context, req *postpb.GetLikesCountRequest) (*postpb.GetLikesCountResponse, error) {
	postID := req.PostId // Assuming postId is passed in the request

	// Call the GetLikeCount service method
	likeCount, err := h.PostService.GetLikeCount(int(postID))
	if err != nil {
		log.Printf("Failed to get like count for post ID %d: %v", postID, err)
		return nil, fmt.Errorf("failed to retrieve like count: %v", err)
	}

	// Prepare the response
	response := &postpb.GetLikesCountResponse{
		LikeCount: int32(likeCount), // Convert to int32 for the response
	}

	return response, nil
}
