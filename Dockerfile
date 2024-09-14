# Use the official Golang image as a base
FROM golang:1.23 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Copy the .env file into the container
COPY .env .env

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/news-feed/main.go

# Use a minimal base image with common utilities
FROM alpine:latest

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main /main

# Copy the .env file from the builder stage
COPY --from=builder /app/.env .env

# Ensure that the binary is executable
RUN chmod +x /main

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["/main"]
