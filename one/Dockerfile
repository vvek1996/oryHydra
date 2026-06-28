# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# Run stage
FROM alpine:3.20

WORKDIR /app

# Copy built binary from the build stage
COPY --from=builder /app/server .

# Set default port environment variable
ENV PORT=4000

# Expose port 4000
EXPOSE 4000

# Run the binary
CMD ["./server"]
