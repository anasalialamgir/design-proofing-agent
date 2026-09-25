FROM golang:1.21-alpine AS builder
WORKDIR /app

# Copy all files from repository
COPY . .

# 1. Automatically create go.mod if it is missing
RUN if [ ! -f go.mod ]; then go mod init design-proofing-agent; fi

# 2. Build from backend/ or root automatically
RUN if [ -f ./backend/main.go ]; then \
        CGO_ENABLED=0 GOOS=linux go build -o /main ./backend; \
    else \
        CGO_ENABLED=0 GOOS=linux go build -o /main .; \
    fi

# Final lightweight runner image
FROM alpine:latest
WORKDIR /app
COPY --from=builder /main .

EXPOSE 8080
ENV PORT=8080

CMD ["./main"]
