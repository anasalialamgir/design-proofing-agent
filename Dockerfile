FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download || true
COPY backend/ ./backend/
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./backend

FROM alpine:latest
RUN apk --no-cache add ca-certificates font-noto
WORKDIR /root/
COPY --from=builder /server .
EXPOSE 8080
CMD ["./server"]
