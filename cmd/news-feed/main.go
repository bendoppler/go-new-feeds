package news_feed

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
	friendService := serviceFactory.CreateFriendsService(friendRepo, postRepo)
	friendsHandler := handlerFactory.CreateFriendsHandler(friendService)
	newsFeedService := serviceFactory.CreateNewsFeedService(postRepo)
	newsFeedHandler := handlerFactory.CreateNewsFeedHandler(newsFeedService)

	http.HandleFunc("/v1/users/login/", userHandler.Login())
	http.HandleFunc("/v1/users/", userHandler.UserHandler)
	http.HandleFunc("/v1/newsfeed/", newsFeedHandler.GetNewsfeed())
	http.HandleFunc("/v1/posts/", postHandler.PostHandler)
	http.HandleFunc("/v1/friends/", friendsHandler.FriendsHandler)

	// Start the server
	addr := ":" + cfg.AppPort
	logger.LogInfo(fmt.Sprintf("Starting server on %s", addr))
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		logger.LogError(err.Error())
	}
}
