package handler

import (
	"news-feed/internal/api/generated/news-feed/friendspb"
	"news-feed/internal/api/generated/news-feed/newsfeedpb"
	"news-feed/internal/api/generated/news-feed/postpb"
	"news-feed/internal/api/generated/news-feed/userpb"
	"news-feed/internal/service"
)

type HandlerFactoryInterface interface {
	CreateUserHandler(userService service.UserServiceInterface) UserHandlerInterface
	CreatePostHandler(postService service.PostServiceInterface) PostHandlerInterface
	CreateFriendsHandler(friendsService service.FriendsServiceInterface) FriendsHandlerInterface
	CreateNewsFeedHandler(newsFeedService service.NewsFeedServiceInterface) NewsFeedHandlerInterface
}

type HandlerFactory struct{}

func (*HandlerFactory) CreateUserHandler(userService userpb.UserServiceClient) UserHandlerInterface {
	return &UserHandler{
		grpcUserHandler: userService,
	}
}

func (*HandlerFactory) CreatePostHandler(postService postpb.PostServiceClient) PostHandlerInterface {
	return &PostHandler{grpcPostHandler: postService}
}

func (*HandlerFactory) CreateFriendsHandler(friendsService friendspb.FriendsServiceClient) FriendsHandlerInterface {
	return &FriendsHandler{grpcFriendsHandler: friendsService}
}

func (*HandlerFactory) CreateNewsFeedHandler(newsFeedService newsfeedpb.NewsfeedServiceClient) NewsFeedHandlerInterface {
	return &NewsfeedHandler{newsFeedService: newsFeedService}
}
