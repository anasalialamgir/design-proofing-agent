FROM golang:1.21-alpine AS builder
WORKDIR /build

# Copy all files from repository into a temporary folder
COPY . /src

# Create a clean go.mod
RUN go mod init design-proofing-agent

# Locate and isolate ONLY main.go
RUN if [ -f /src/backend/main.go ]; then \
        cp /src/backend/main.go ./main.go; \
    elif [ -f /src/main.go ]; then \
        cp /src/main.go ./main.go; \
    else \
        find /src -name "main.go" -exec cp {} ./main.go \; ; \
    fi

# Build ONLY the self-contained main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /main ./main.go

# Final lightweight runner
FROM alpine:latest
WORKDIR /app
COPY --from=builder /main .

EXPOSE 8080
ENV PORT=8080

CMD ["./main"]
