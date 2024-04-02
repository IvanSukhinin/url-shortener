FROM golang:alpine
WORKDIR /app
EXPOSE 8090
CMD go run cmd/url-shortener/main.go
