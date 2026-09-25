FROM golang:1.21-alpine AS builder
WORKDIR /app

# Copy all files
COPY . .

# Initialize Go module if not present
RUN if [ ! -f go.mod ]; then go mod init design-proofing-agent; fi

# Build the main.go application directly
RUN CGO_ENABLED=0 GOOS=linux go build -o /main main.go

# Production runner image
FROM alpine:latest
WORKDIR /app
COPY --from=builder /main .

EXPOSE 8080
ENV PORT=8080

CMD ["./main"]
