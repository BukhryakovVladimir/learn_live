# Start from the official Golang image
FROM golang:latest as builder

# Add Maintainer info
LABEL maintainer="Vladimir Bukhryakov <buhryakov15@mail.ru>"

# Install git.
# Git is required for fetching the dependencies.
RUN apt-get update && apt-get install -y git

# Set the current working directory inside the container
WORKDIR /go/src/app/

# Copy the source from the current directory to the working Directory inside the container
COPY ./cmd/learn_live /go/src/app/cmd/learn_live
COPY ./internal /go/src/app/internal

# Copy go mod and sum files
COPY go.mod go.sum /go/src/app/

# Copy the environment file
COPY .env /go/src/app/.env

# Download all dependencies. Dependencies will be cached if the go.mod and the go.sum files are not changed
RUN go mod download

WORKDIR /go/src/app/cmd/learn_live

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Start a new stage from scratch
FROM golang:latest
RUN apt-get update && apt-get install -y ca-certificates

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage. Observe we also copied the .env file
COPY --from=builder /go/src/app/cmd/learn_live/main .
COPY --from=builder /go/src/app/.env .

# Copy SSL certificates
# COPY ssl /etc/golang/ssl

# Expose port 3000 to the outside world
EXPOSE 3000

# Install dockerize
RUN apt-get install -y wget
RUN wget https://github.com/jwilder/dockerize/releases/download/v0.6.1/dockerize-linux-amd64-v0.6.1.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-linux-amd64-v0.6.1.tar.gz \
    && rm dockerize-linux-amd64-v0.6.1.tar.gz

# Wait for PostgreSQL to be ready before running the application
CMD dockerize -wait tcp://learn_live-postgres:5432 -timeout 60s ./main
