# syntax=docker/dockerfile:1
# latest Dockerfile syntax

FROM golang:1.23-alpine

WORKDIR /backend

# Install build dependencies
## musl-dev included to support glibc dependencies
## (https://stackoverflow.com/questions/52056387/how-to-install-go-in-alpine-linux)
RUN apk add --no-cache gcc musl-dev

# Create separate caching layer for Go dependencies. If these files have
# not changed, Docker will use the cache for dependency resolution even 
# if other source code has changed
# (avoid re-downloading unchanged modules)
COPY go.mod go.sum ./

# Download and cache dependencies
# can use default Go module storage location since this is the only Docker build
# created by the test runner, so there will be no conflicts with caches from
# other builds
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy other source code in new cache layer separate from dependencies
COPY . .

# Run tests
CMD ["go", "test", "--tags", "fts5", "-v", "./..."]