# Stage 1: Build the project two validator CLI
FROM golang:1.23-alpine AS builder-validator
WORKDIR /build
COPY two/go.mod ./
# Copy CLI source code
COPY two/ .
# Build statically linked binary for Linux
RUN CGO_ENABLED=0 GOOS=linux go build -o test1 .

# Stage 2: Build the project one web server
FROM golang:1.23-alpine AS builder-server
WORKDIR /build
COPY one/go.mod one/go.sum ./
RUN go mod download
# Copy server source code
COPY one/ .
# Build statically linked binary for Linux
RUN CGO_ENABLED=0 GOOS=linux go build -o server .



# Stage 3: Final runtime stage
FROM alpine:3.20
WORKDIR /app

# Install git for repository cloning/pulling
RUN apk add --no-cache git

# Copy binaries from the build stages
COPY --from=builder-server /build/server .
COPY --from=builder-validator /build/test1 .
# Configure path to the validator CLI and default port
ENV VALIDATOR_PATH=/app/test1
ENV PORT=4000

# Expose port
EXPOSE 4000

# Run the web server
CMD ["./server"]
