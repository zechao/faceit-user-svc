# Build stage
FROM golang:1.24-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Install necessary packages
RUN apk add --no-cache git

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download the Go modules
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the application
RUN go build -o main ./cmd/main.go

# Final stage
FROM alpine:latest

# Install necessary packages
RUN apk add --no-cache ca-certificates

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary and migration files from the builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./main"]