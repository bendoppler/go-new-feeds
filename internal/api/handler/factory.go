package handler

import "news-feed/internal/service"

type HandlerFactoryInterface interface {
	CreateUserHandler(userService service.UserServiceInterface) UserHandlerInterface
	CreatePostHandler(postService service.PostService) PostHandlerInterface
}

type HandlerFactory struct{}

func (*HandlerFactory) CreateUserHandler(userService service.UserServiceInterface) UserHandlerInterface {
	return &UserHandler{
		userService: userService,
	}
}

func (*HandlerFactory) CreatePostHandler(postService service.PostService) PostHandlerInterface {
	return &PostHandler{postService: postService}
}
