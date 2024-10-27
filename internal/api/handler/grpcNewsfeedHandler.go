package handler

import (
	"context"
	"github.com/golang/protobuf/ptypes"
	"log"
	"news-feed/internal/api/generated/news-feed/newsfeedpb"
	"news-feed/internal/service"
)

type GRPCNewsfeedHandler struct {
	newsfeedpb.UnimplementedNewsfeedServiceServer
	newsFeedService service.NewsFeedServiceInterface
}

func NewNewsfeedHandler(newsFeedService service.NewsFeedServiceInterface) *GRPCNewsfeedHandler {
	return &GRPCNewsfeedHandler{newsFeedService: newsFeedService}
}

func (h *GRPCNewsfeedHandler) GetNewsfeed(ctx context.Context, req *newsfeedpb.GetNewsfeedRequest) (*newsfeedpb.GetNewsfeedResponse, error) {
	posts, err := h.newsFeedService.GetNewsfeedPosts()
	if err != nil {
		log.Printf("Failed to get newsfeed posts: %v", err)
		return nil, err
	}

	// Map posts to the gRPC response format
	var responsePosts []*newsfeedpb.Post
	for _, post := range posts {
		createdAtProto, err := ptypes.TimestampProto(post.CreatedAt)
		if err != nil {
			log.Printf("Error converting timestamp: %v", err)
			return nil, err
		}

		responsePosts = append(
			responsePosts, &newsfeedpb.Post{
				Id:               int32(post.ID),
				UserId:           int32(post.UserID),
				ContentText:      post.ContentText,
				ContentImagePath: post.ContentImagePath,
				CreatedAt:        createdAtProto,
			},
		)
	}

	return &newsfeedpb.GetNewsfeedResponse{Posts: responsePosts}, nil
}
