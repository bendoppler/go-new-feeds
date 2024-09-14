package handler

import "news-feed/internal/service"

type HandlerFactoryInterface interface {
	CreateUserHandler(userService service.UserServiceInterface) UserHandlerInterface
	CreatePostHandler(postService service.PostServiceInterface) PostHandlerInterface
	CreateFriendsHandler(friendsService service.FriendsServiceInterface) FriendsHandlerInterface
	CreateNewsFeedHandler(newsFeedService service.NewsFeedServiceInterface) NewsFeedHandlerInterface
}

type HandlerFactory struct{}

func (*HandlerFactory) CreateUserHandler(userService service.UserServiceInterface) UserHandlerInterface {
	return &UserHandler{
		userService: userService,
	}
}

func (*HandlerFactory) CreatePostHandler(postService service.PostServiceInterface) PostHandlerInterface {
	return &PostHandler{postService: postService}
}

func (*HandlerFactory) CreateFriendsHandler(friendsService service.FriendsServiceInterface) FriendsHandlerInterface {
	return &FriendsHandler{friendsService: friendsService}
}

func (*HandlerFactory) CreateNewsFeedHandler(newsFeedService service.NewsFeedServiceInterface) NewsFeedHandlerInterface {
	return &NewsfeedHandler{newsFeedService: newsFeedService}
}
