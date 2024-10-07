package main

import (
	"fmt"
	"net/http"
	"news-feed/internal/api/handler"
	"news-feed/internal/db"
	"news-feed/internal/repository"
	"news-feed/internal/service"
	"news-feed/internal/storage"
	"news-feed/pkg/config"
	"news-feed/pkg/logger"
	"news-feed/pkg/middleware"
	"time"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig(".env.userPostFriends")

	// Initialize logger
	logger.InitLogger()

	// Initialize the database connection
	factory := db.PersistentFactory{}
	mySQLDB, err := factory.CreateMySQLDatabase()
	repositoryFactory := &repository.RepositoryFactory{}
	serviceFactory := &service.ServiceFactory{}
	handlerFactory := &handler.HandlerFactory{}
	storageFactory := &storage.StorageFactory{}
	minioStorage, err := storageFactory.CreateMinioStorage()
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to create minio storage: %v", err))
		return
	}

	// Initialize repositories and services
	userRepo := repositoryFactory.CreateUserRepository(mySQLDB)
	userService := serviceFactory.CreateUserService(userRepo)
	userHandler := handlerFactory.CreateUserHandler(userService)
	postRepo := repositoryFactory.CreatePostRepository(mySQLDB)
	postService := serviceFactory.CreatePostService(postRepo, minioStorage, userService)
	postHandler := handlerFactory.CreatePostHandler(postService)
	friendRepo := repositoryFactory.CreateFriendRepository(mySQLDB)
	friendService := serviceFactory.CreateFriendsService(friendRepo, postRepo, userRepo)
	friendsHandler := handlerFactory.CreateFriendsHandler(friendService)

	go userService.PeriodicallyRefreshBloomFilter(1 * time.Hour)

	// Populate the Bloom filter
	err = userService.InitializeBloomFilter()
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to initialize Bloom Filter: %v", err))
		return
	}

	// Routes

	// @Summary User login
	// @Description Login user and return access token.
	// @Tags Users
	// @Accept  json
	// @Produce  json
	// @Param   user  body     handler.LoginRequest true "User credentials"
	// @Success 200 {object} handler.LoginResponse
	// @Failure 400 {object} handler.ErrorResponse
	// @Router /v1/users/login [post]
	http.HandleFunc("/v1/users/login", userHandler.Login())

	// @Summary Get user information
	// @Description Retrieve user information by ID.
	// @Tags Users
	// @Produce  json
	// @Param   id     path     int     true  "User ID"
	// @Success 200 {object} entity.User
	// @Failure 404 {object} handler.ErrorResponse
	// @Router /v1/users [get]
	http.HandleFunc("/v1/users", userHandler.UserHandler)

	// @Summary Create post
	// @Description Create a new post.
	// @Tags Posts
	// @Accept  json
	// @Produce  json
	// @Param   post   body      entity.Post  true  "New post details"
	// @Success 201 {object} entity.Post
	// @Failure 400 {object} handler.ErrorResponse
	// @Router /v1/posts [post]
	http.HandleFunc("/v1/posts", middleware.JWTAuthMiddleware(postHandler.CreatePost()).ServeHTTP)

	// @Summary Get post by ID
	// @Description Retrieve post details by post ID.
	// @Tags Posts
	// @Produce  json
	// @Param   id   path      int  true  "Post ID"
	// @Success 200 {object} entity.Post
	// @Failure 404 {object} handler.ErrorResponse
	// @Router /v1/posts/{id} [get]
	http.HandleFunc("/v1/posts/", postHandler.PostHandler)

	// @Summary Manage friends
	// @Description Manage friend relationships.
	// @Tags Friends
	// @Produce  json
	// @Param   id   path      int  true  "Friend ID"
	// @Success 200 {object} entity.Friend
	// @Failure 404 {object} handler.ErrorResponse
	// @Router /v1/friends/{id} [get]
	http.HandleFunc("/v1/friends/", friendsHandler.FriendsHandler)

	// Start the server
	addr := ":" + cfg.AppPort
	logger.LogInfo(fmt.Sprintf("Starting User-Post-Friends service on %s", addr))
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		logger.LogError(err.Error())
		return
	}
}
