# Stage 1: Build the Go web servers
FROM golang:1.23-alpine AS builder
WORKDIR /build
COPY go.mod ./
RUN go mod download
# Copy the source code
COPY . .
# Build statically linked binaries for Linux
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o second ./cmd/second/main.go

# Stage 2: Final runtime stage
FROM alpine:3.20
WORKDIR /app
COPY --from=builder /build/server .
COPY --from=builder /build/second .
EXPOSE 8080 8081
CMD ["./server"]
