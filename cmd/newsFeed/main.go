package main

import (
	"fmt"
	"google.golang.org/grpc"
	"net"
	"news-feed/internal/api/generated/news-feed/newsfeedpb"
	"news-feed/internal/api/handler"
	"news-feed/internal/db"
	"news-feed/internal/repository"
	"news-feed/internal/service"
	"news-feed/pkg/config/newsfeed"
	"news-feed/pkg/logger"
)

func main() {
	// Load configuration
	cfg := newsfeed.LoadNewsfeedConfig()

	// Initialize logger
	logger.InitLogger()

	// Initialize the database connection
	factory := db.PersistentFactory{}
	mySQLDB, err := factory.CreateMySQLDatabase()

	// Initialize repositories and services
	repositoryFactory := &repository.RepositoryFactory{}
	serviceFactory := &service.ServiceFactory{}

	postRepo := repositoryFactory.CreatePostRepository(mySQLDB) // Provide necessary db connection
	newsFeedService := serviceFactory.CreateNewsFeedService(postRepo)
	newsFeedHandler := handler.NewNewsfeedHandler(newsFeedService)

	// Set up gRPC server
	grpcServer := grpc.NewServer()
	newsfeedpb.RegisterNewsfeedServiceServer(grpcServer, newsFeedHandler)

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
