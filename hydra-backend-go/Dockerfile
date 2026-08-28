FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 4000

CMD ["./main"]
