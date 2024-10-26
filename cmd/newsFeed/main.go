package main

import (
	"fmt"
	"net/http"
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
	handlerFactory := &handler.HandlerFactory{}

	postRepo := repositoryFactory.CreatePostRepository(mySQLDB) // Provide necessary db connection
	newsFeedService := serviceFactory.CreateNewsFeedService(postRepo)
	newsFeedHandler := handlerFactory.CreateNewsFeedHandler(newsFeedService)

	// @Summary Get news feed
	// @Description Get the latest posts from user's friends.
	// @Tags NewsFeed
	// @Produce  json
	// @Param   cursor   query    string  false  "Pagination cursor"
	// @Param   limit    query    int     false  "Limit"
	// @Success 200 {array} entity.Post
	// @Failure 404 {object} handler.ErrorResponse
	// @Router /v1/newsfeed [get]
	http.HandleFunc("/v1/newsfeed", newsFeedHandler.GetNewsfeed())

	// Start the server
	addr := ":" + cfg.AppPort
	logger.LogInfo(fmt.Sprintf("Starting Newsfeed service on %s", addr))
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		logger.LogError(err.Error())
		return
	}
}
