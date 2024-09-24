package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
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
	cfg := config.LoadConfig()

	// Initialize logger
	logger.InitLogger()
	// Initialize the database connection
	factory := db.PersistentFactory{}
	mySQLDB, err := factory.CreateMySQLDatabase()
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to connect to database: %v", err))
		return
	}

	repositoryFactory := &repository.RepositoryFactory{}
	serviceFactory := &service.ServiceFactory{}
	handlerFactory := &handler.HandlerFactory{}
	storageFactory := &storage.StorageFactory{}

	minioStorage, err := storageFactory.CreateMinioStorage()
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to create minio storage: %v", err))
		return
	}

	userRepo := repositoryFactory.CreateUserRepository(mySQLDB)
	userService := serviceFactory.CreateUserService(userRepo)
	userHandler := handlerFactory.CreateUserHandler(userService)
	postRepo := repositoryFactory.CreatePostRepository(mySQLDB)
	postService := serviceFactory.CreatePostService(postRepo, minioStorage)
	postHandler := handlerFactory.CreatePostHandler(postService)
	friendRepo := repositoryFactory.CreateFriendRepository(mySQLDB)
	friendService := serviceFactory.CreateFriendsService(friendRepo, postRepo, userRepo)
	friendsHandler := handlerFactory.CreateFriendsHandler(friendService)
	newsFeedService := serviceFactory.CreateNewsFeedService(postRepo)
	newsFeedHandler := handlerFactory.CreateNewsFeedHandler(newsFeedService)

	go func() {
		logger.LogInfo(fmt.Sprintf("Attempting to start pprof server"))
		err := http.ListenAndServe("localhost:6060", nil)
		if err != nil {
			logger.LogError(fmt.Sprintf("Error when run pprof: %v", err))
			return
		}
		logger.LogInfo(fmt.Sprintf("Starting pprof server on localhost:6060"))
	}()
	go userService.PeriodicallyRefreshBloomFilter(1 * time.Hour)

	// Populate the Bloom filter
	err = userService.InitializeBloomFilter()
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to initialize Bloom Filter: %v", err))
		return
	}

	http.HandleFunc("/v1/users/login", userHandler.Login())
	http.HandleFunc("/v1/users", userHandler.UserHandler)
	http.HandleFunc("/v1/newsfeed", newsFeedHandler.GetNewsfeed())
	http.HandleFunc("/v1/posts", middleware.JWTAuthMiddleware(postHandler.CreatePost()).ServeHTTP)
	http.HandleFunc("v1/posts/", postHandler.PostHandler)
	http.HandleFunc("/v1/friends/", friendsHandler.FriendsHandler)

	// Catch-all handler for unhandled endpoints
	http.HandleFunc(
		"/", func(w http.ResponseWriter, r *http.Request) {
			// Log the unhandled endpoint
			logger.LogWarning(fmt.Sprintf("Unhandled endpoint: %s %s", r.Method, r.URL.Path))
			// Return a 404 Not Found response
			http.Error(w, "404 Not Found: Endpoint does not exist", http.StatusNotFound)
		},
	)

	// Start the server
	addr := ":" + cfg.AppPort
	logger.LogInfo(fmt.Sprintf("Starting server on %s", addr))
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		logger.LogError(err.Error())
		return
	}
}
