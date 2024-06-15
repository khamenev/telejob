# Use an official Golang image as a base
FROM golang:1.20 as builder

# Установите необходимые пакеты для сборки TDLib
RUN apt-get update && apt-get install -y \
    cmake \
    g++ \
    zlib1g-dev \
    libssl-dev \
    git

# Клонируйте и соберите TDLib
WORKDIR /tdlib
RUN git clone https://github.com/tdlib/td.git .
RUN mkdir build && cd build && \
    cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX:PATH=/usr/local .. && \
    cmake --build . --target install


# Set the Current Working Directory inside the container
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

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /usr/local/lib/libtd* /usr/local/lib/

# Install cron and necessary packages
RUN apk add --no-cache bash curl tzdata postgresql-client busybox-suid ca-certificates && \
    update-ca-certificates

 RUN apk add --no-cache libstdc++ libgcc libssl1.1 ca-certificates && \
     update-ca-certificates

RUN echo '/usr/local/lib' >> /etc/ld.so.conf.d/local.conf && ldconfig

# Copy entrypoint script
COPY entrypoint.sh /root/entrypoint.sh

# Make the entrypoint script executable
RUN chmod +x /root/entrypoint.sh

# Run the entrypoint script
ENTRYPOINT ["/bin/sh", "/root/entrypoint.sh"]
