# Use an official Golang image as a base
FROM golang:1.22-alpine

# Set the Current Working Directory inside the container
WORKDIR /root/telejob

# Install cron and necessary packages
RUN apk update && \
    apk add --no-cache bash curl tzdata openssl-dev postgresql-dev alpine-sdk linux-headers git zlib-dev openssl-dev gperf php cmake

RUN git clone --depth 1 --branch v1.8.0 https://github.com/tdlib/td.git \
    && cd td \
    && rm -rf build \
    && mkdir build \
    && cd build \
    && cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX:PATH=/usr/local .. \
    && cmake --build . --target install -- -j10 \
    && ls -la /usr/local/lib | grep json


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
