# Build Stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the binary
# CGO_ENABLED=0 ensures a static binary
RUN CGO_ENABLED=0 GOOS=linux go build -o camera-observer cmd/camera-observer/main.go

# Run Stage
FROM alpine:latest

WORKDIR /root/

# Install certificates for HTTPS (SmartThings API)
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/camera-observer .
COPY config.yaml . 

# Expose port? Not needed as it's a client, but good practice if we add metrics later.
# EXPOSE 8080

CMD ["./camera-observer"]
