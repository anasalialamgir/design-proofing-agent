# Stage 1: Build the Go application
FROM golang:1.21-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum (if they exist, otherwise copy all)
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./main.go

# Stage 2: Create a lightweight runtime image
FROM alpine:latest

WORKDIR /app

# Copy the pre-built binary from the builder stage
COPY --from=builder /app/main .

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
