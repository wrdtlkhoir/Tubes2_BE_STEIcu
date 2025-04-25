# syntax=docker/dockerfile:1
FROM golang:1.21-alpine AS build
WORKDIR /app
COPY src/go.mod src/go.sum ./
RUN go mod download
COPY src/ .
RUN go build -o server

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/server .
COPY src/recipes.json .
EXPOSE 8080
CMD ["./server"]
