FROM golang:1.24-alpine3.20 

# Install necessary build tools and SQLite headers
RUN apk update && apk add --no-cache \
  build-base \
  sqlite-dev \
  sqlite \
  gcc \
  musl-dev

# Enable CGO
ENV CGO_ENABLED=1
