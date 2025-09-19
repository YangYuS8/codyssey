# Go backend Dockerfile
FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server ./

FROM alpine:3.19
WORKDIR /app
COPY --from=build /app/server /app/server
EXPOSE 8080
ENV GO_BACKEND_PORT=8080
ENTRYPOINT ["/app/server"]
