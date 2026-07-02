# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy dependency manifests
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gopams ./cmd/server/main.go

# Production stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/gopams .

# Copy static frontend files (needed for serving UI assets)
COPY --from=builder /app/frontend ./frontend

# Expose server port
EXPOSE 8080

# Command to run the executable
CMD ["./gopams"]
