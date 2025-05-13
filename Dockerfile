# syntax=docker/dockerfile:1
FROM golang:1.24-alpine
WORKDIR /app/src
COPY src/go.mod src/go.sum ./
RUN go mod download
COPY src/ ./

RUN go build -o main
EXPOSE 8080
CMD ["./main"]