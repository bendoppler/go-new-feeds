package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc"
	"net/http"
	_ "net/http/pprof"
	_ "news-feed/docs"
	"news-feed/internal/api/generated/news-feed/friendspb"
	"news-feed/internal/api/generated/news-feed/newsfeedpb"
	"news-feed/internal/api/generated/news-feed/postpb"
	"news-feed/internal/api/generated/news-feed/userpb"
	"news-feed/internal/api/handler"
	"news-feed/pkg/config/webApp"
	"news-feed/pkg/logger"
	"news-feed/pkg/middleware"
)

// @title News Feed API
// @version 1.0
// @description This is a sample news feed server.
// @termsOfService http://news-feed.com/terms/

// @contact.name API Support
// @contact.url http://news-feed.com/support
// @contact.email support@news-feed.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /v1

func main() {
	// Load configuration
	cfg := webApp.LoadConfig()

	// Initialize logger
	logger.InitLogger()

	conn, err := grpc.Dial(
		"localhost:"+cfg.PostUserFriendsPort, grpc.WithInsecure(),
	) // Use secure connection in production
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to connect to post user service server: %v", err))
		return
	}
	handlerFactory := &handler.HandlerFactory{}
	userService := userpb.NewUserServiceClient(conn)
	userHandler := handlerFactory.CreateUserHandler(userService)
	postService := postpb.NewPostServiceClient(conn)
	postHandler := handlerFactory.CreatePostHandler(postService)
	friendsService := friendspb.NewFriendsServiceClient(conn)
	friendsHandler := handlerFactory.CreateFriendsHandler(friendsService)
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to close connection to post user service server: %v", err))
		}
	}(conn)

	newsfeedConn, err := grpc.Dial("localhost:"+cfg.NewsfeedAppPort, grpc.WithInsecure())
	newsfeedService := newsfeedpb.NewNewsfeedServiceClient(newsfeedConn)
	newsfeedHandler := handlerFactory.CreateNewsFeedHandler(newsfeedService)
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to connect news feed service server: %v", err))
		return
	}
	defer func(newsfeedConn *grpc.ClientConn) {
		err := newsfeedConn.Close()
		if err != nil {
			logger.LogError(fmt.Sprintf("Failed to close connection to newsfeed service server: %v", err))
		}
	}(newsfeedConn)

	go func() {
		logger.LogInfo(fmt.Sprintf("Attempting to start pprof server"))
		err := http.ListenAndServe("0.0.0.0:6060", nil)
		if err != nil {
			logger.LogError(fmt.Sprintf("Error when run pprof: %v", err))
			return
		}
		logger.LogInfo(fmt.Sprintf("Starting pprof server on :6060"))
	}()

	// Prometheus metrics
	http.Handle("/metrics", promhttp.Handler())

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

	// @Summary Get news feed
	// @Description Get the latest posts from user's friends.
	// @Tags NewsFeed
	// @Produce  json
	// @Param   cursor   query    string  false  "Pagination cursor"
	// @Param   limit    query    int     false  "Limit"
	// @Success 200 {array} entity.Post
	// @Failure 404 {object} handler.ErrorResponse
	// @Router /v1/newsfeed [get]
	http.HandleFunc("/v1/newsfeed", newsfeedHandler.GetNewsfeed())

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

	// Swagger documentation route
	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

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
