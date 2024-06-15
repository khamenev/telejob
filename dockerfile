# Use an official Golang image as a base
FROM golang:1.20 as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Install necessary dependencies for TDLib and OpenSSH
RUN apt-get update && apt-get install -y \
    cmake \
    g++ \
    zlib1g-dev \
    libssl-dev \
    git \
    gperf \
    openssh-server

# Clone and build TDLib
WORKDIR /tdlib
RUN git clone https://github.com/tdlib/td.git .
RUN mkdir build && cd build && \
    cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX:PATH=/usr/local .. && \
    cmake --build . --target install

# Set the Current Working Directory back to the application directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o main .

# Use a minimal image to run the application
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /root/

# Install necessary packages
RUN apk add --no-cache bash curl tzdata postgresql-client busybox-suid ca-certificates && \
    update-ca-certificates

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /usr/local/lib/libtdjson.so /usr/lib

# Configure SSH
RUN mkdir /var/run/sshd && echo 'root:password' | chpasswd && sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && sed -i 's/UsePAM yes/UsePAM no/' /etc/ssh/sshd_config

# Copy entrypoint script
COPY entrypoint.sh /root/entrypoint.sh

# Make the entrypoint script executable
RUN chmod +x /root/entrypoint.sh

# Run the entrypoint script
ENTRYPOINT ["/bin/sh", "/root/entrypoint.sh"]
