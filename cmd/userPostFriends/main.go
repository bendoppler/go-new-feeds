package main

import (
	"fmt"
	"google.golang.org/grpc"
	"net"
	"news-feed/internal/api/generated/news-feed/friendspb"
	"news-feed/internal/api/generated/news-feed/postpb"
	"news-feed/internal/api/generated/news-feed/userpb"
	"news-feed/internal/api/handler"
	"news-feed/internal/db"
	"news-feed/internal/repository"
	"news-feed/internal/service"
	"news-feed/internal/storage"
	"news-feed/pkg/config/userPostFriends"
	"news-feed/pkg/logger"
	"time"
)

func main() {
	// Load configuration
	cfg := userPostFriends.LoadUserPostFriendsConfig()

	// Initialize logger
	logger.InitLogger()

	// Initialize the database connection
	factory := db.PersistentFactory{}
	mySQLDB, err := factory.CreateMySQLDatabase()
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to create mySQLDB: %v", err))
		return
	}
	repositoryFactory := &repository.RepositoryFactory{}
	serviceFactory := &service.ServiceFactory{}
	storageFactory := &storage.StorageFactory{}

	minioStorage, err := storageFactory.CreateMinioStorage()
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to create minio storage: %v", err))
		return
	}

	// Initialize repositories and services
	userRepo := repositoryFactory.CreateUserRepository(mySQLDB)
	userService := serviceFactory.CreateUserService(userRepo)
	userHandler := handler.GRPCUserHandler{
		UserService: userService,
	}
	postRepo := repositoryFactory.CreatePostRepository(mySQLDB)
	postService := serviceFactory.CreatePostService(postRepo, minioStorage, userService)
	postHandler := handler.GRPCPostHandler{
		PostService: postService,
	}
	friendRepo := repositoryFactory.CreateFriendRepository(mySQLDB)
	friendService := serviceFactory.CreateFriendsService(friendRepo, postRepo, userRepo)
	friendsHandler := handler.GRPCFriendsHandler{
		FriendsService: friendService,
	}

	go userService.PeriodicallyRefreshBloomFilter(1 * time.Hour)

	// Populate the Bloom filter
	err = userService.InitializeBloomFilter()
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to initialize Bloom Filter: %v", err))
		return
	}

	// Set up gRPC server
	grpcServer := grpc.NewServer()
	// Assuming `RegisterUserServiceServer` is generated by protobuf for your `UserService`
	userpb.RegisterUserServiceServer(grpcServer, &userHandler)
	postpb.RegisterPostServiceServer(grpcServer, &postHandler)
	friendspb.RegisterFriendsServiceServer(grpcServer, &friendsHandler)

	// Start listening on the configured port
	port := cfg.AppPort
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to listen on port %s: %v", port, err))
		return
	}
	if err := grpcServer.Serve(lis); err != nil {
		logger.LogError(fmt.Sprintf("Failed to serve gRPC server: %v", err))
	}
}
