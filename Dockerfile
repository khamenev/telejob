# Use an official Golang image as a base
FROM golang:1.22-alpine

# Set the Current Working Directory inside the container
WORKDIR /root/telejob

# Install cron and necessary packages
RUN apk add --no-cache telegram-tdlib bash curl tzdata libgcc libssl1.1 postgresql-client ca-certificates && \
    update-ca-certificates

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o main .

# Copy entrypoint script
COPY entrypoint.sh /root/entrypoint.sh

# Make the entrypoint script executable
RUN chmod +x /root/entrypoint.sh

# Run the entrypoint script
ENTRYPOINT ["/bin/sh", "/root/entrypoint.sh"]
