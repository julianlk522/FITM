FROM golang:1.23-alpine

WORKDIR /backend

# Install build dependencies
## musl-dev included to support glibc dependencies
## (https://stackoverflow.com/questions/52056387/how-to-install-go-in-alpine-linux)
RUN apk add --no-cache gcc musl-dev

# Run tests
CMD ["go", "test", "--tags", "fts5", "-v", "./..."]