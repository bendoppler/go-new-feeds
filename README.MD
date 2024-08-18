# Project Structure

This project follows the [golang-standards/project-layout](https://github.com/golang-standards/project-layout) guidelines. Below is the structure of the project and an explanation of each folder and file.

## Directory Structure

```plaintext
go-newsfeed-server/
├── cmd/
│   └── news-feed/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── handler/
│   │   │   ├── factory.go
│   │   │   ├── friendsHandler.go
│   │   │   ├── newsFeedHandler.go
│   │   │   ├── postHandler.go
│   │   │   └── userHandler.go
│   ├── cache/
│   │   └── redisCache.go
│   ├── db/
│   │   ├── mysqlDB.go
│   │   └── migration.go
│   ├── entity/
│   │   ├── post.go
│   │   ├── comment.go
│   │   ├── follower.go
│   │   ├── user.go
│   │   └── like.go
│   ├── repository/
│   │   ├── friendRepository.go
│   │   ├── postRepository.go
│   │   └── userRepository.go
│   ├── service/
│   │   ├── factory.go
│   │   ├── friendService.go
│   │   ├── newsFeedService.go
│   │   ├── postService.go
│   │   └── userService.go
│   └── storage/
│       └── minioStorage.go
├── config/
│   └── config.go
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── go.sum
```
## Directory Details

### `cmd/`
Contains the main application entry point.
- **`news-feed/`**
    - **`main.go`**: Initializes and runs the application.

### `internal/`
Contains application code that is not intended to be used outside of this project.
- **`api/`**
    - **`handler/`**
        - **`friends.go`**: Handles friend-related API endpoints.
        - **`newsfeeds.go`**: Handles newsfeed-related API endpoints.
        - **`posts.go`**: Handles post-related API endpoints.
        - **`users.go`**: Handles user-related API endpoints.
- **`cache/`**
    - **`redis_cache.go`**: Implements caching using Redis.
- **`db/`**
    - **`db.go`**: Database configuration and connection.
    - **`migration.go`**: Database migration scripts.
- **`entity/`**
    - **`like.go`**: Defines the Like entity.
    - **`post.go`**: Defines the Post entity.
    - **`user.go`**: Defines the User entity.
    - **`new.go`**: Defines the User-User relationship entity.
- **`repository/`**
    - **`friendRepository.go`**: Handles database operations for friend.
    - **`likeRepository.go`**: Handles database operations for likes.
    - **`postRepository.go`**: Handles database operations for posts.
    - **`userRepository.go`**: Handles database operations for users.
- **`service/`**
    - **`friendService.go`**: Business logic for friends.
    - **`newsFeedService.go`**: Business logic for new feeds.
    - **`postService.go`**: Business logic for posts.
    - **`userService.go`**: Business logic for users.
- **`storage/`**
    - **`minioStorage.go`**: Implements file storage using MinIO.

### `config/`
Contains configuration files.
- **`config.go`**: Application configuration settings.

### Root Directory
- **`docker-compose.yml`**: Defines Docker services for the project.
- **`Dockerfile`**: Dockerfile for building the application image.
- **`go.mod`**: Go module file specifying dependencies.
- **`go.sum`**: Go sum file with checksums for dependencies.

## Getting Started

### Prerequisites

Before you begin, ensure you have the following installed on your system:

- [Docker](https://docs.docker.com/get-docker/) - For containerization and managing dependencies.
- [Docker Compose](https://docs.docker.com/compose/install/) - For defining and running multi-container Docker applications.
- [Go](https://golang.org/doc/install) - Go programming language (version specified in `go.mod`).

### Setting Up

1. **Clone the Repository**

   Clone this repository to your local machine using:

   ```bash
   git clone https://github.com/yourusername/go-newsfeed-server.git
   cd go-newsfeed-server
   ```
### Create a `.env` File

Create a `.env` file in the root directory of the project. This file should contain the environment variables required for the application to run. Here's a sample `.env` file:

```env
DB_USER=root
DB_PASSWORD=yourpassword
DB_NAME=yourdatabase
DB_ROOT_PASSWORD=yourrootpassword
REDIS_PASSWORD=yourredispassword
MINIO_ACCESS_KEY=yourminioaccesskey
MINIO_SECRET_KEY=yourminiosecretkey
```
Update the values with your own configuration:

- **`DB_USER`**: MySQL username.
- **`DB_PASSWORD`**: MySQL user password.
- **`DB_NAME`**: MySQL database name.
- **`DB_ROOT_PASSWORD`**: MySQL root password.
- **`REDIS_PASSWORD`**: Redis password.
- **`MINIO_ACCESS_KEY`**: MinIO access key.
- **`MINIO_SECRET_KEY`**: MinIO secret key.

### Build and Run the Docker Application

Use Docker Compose to build and run the containers. This command will:

1. Build the Docker images for your application based on the `Dockerfile`.
2. Start the containers as defined in the `docker-compose.yml` file.

Run the following command in the root of your project directory:

```bash
docker-compose up --build
```
- `--build`: Forces Docker Compose to rebuild the images even if they are up-to-date.
- `--detach` or `-d`: Runs the containers in the background and prints the container IDs.
- `--remove-orphans`: Removes containers for services not defined in the `docker-compose.yml` file.

To build and run the Docker application, execute the following command:

```bash
docker-compose up --build
```
This command will build the Docker images as defined in the `Dockerfile` and `docker-compose.yml` file, then start the containers in the background. If you make changes to the Dockerfile or dependencies, you can re-run this command to rebuild the images and restart the containers.

After running the application, you can access the application at `http://localhost:8080`.

To stop and remove the application, networks, and volumes created by `docker-compose up`, use the following command:

```bash
docker-compose down
```
This command will stop the running application and remove them, along with the networks and volumes that were created. It is useful for cleaning up after development or when you want to ensure a fresh start.

If you need to stop the application without removing them, you can use:

```bash
docker-compose stop
```
This will stop the running application but leave them in place, so they can be restarted later.

To restart the application, you can use:

```bash
docker-compose start
```

This command starts the stopped application without rebuilding the images. It is a quick way to resume your application if you need to pause and resume development or testing.

To remove all application, networks, and volumes defined in your docker-compose.yml, you can use:

```bash
docker-compose down --volumes
```
This command stops and removes the application, networks, and volumes associated with it. It is useful for cleaning up your environment or ensuring a fresh start.

For more detailed information on Docker Compose commands and options, refer to the [official Docker Compose documentation](https://docs.docker.com/compose/).
