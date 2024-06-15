# Use an official Golang image as a base
FROM golang:1.20 as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
WORKDIR /app/cmd
RUN go build -o /main .

# Use a minimal image to run the application
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /root/

# Install necessary packages
RUN apk add --no-cache bash curl tzdata postgresql-client busybox-suid ca-certificates && \
    update-ca-certificates

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /main .
COPY --from=builder /app/migrations ./migrations

# Expose port
EXPOSE 80

# Run the Go app
CMD ["./main"]
