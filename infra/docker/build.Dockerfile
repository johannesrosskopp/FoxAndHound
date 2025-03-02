# Build stage
FROM golang:1.23.5-bullseye AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files first to leverage Docker cache
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application files
COPY . .

# Build the Go binary
RUN go build -o backend-build .

# Final stage
FROM debian:bookworm-slim

# Set the working directory
WORKDIR /app

# Copy the Go binary from the builder stage
COPY --from=builder /app/backend-build .

# Copy the certs folder from the builder stage
COPY --from=builder /app/certs /app/certs

# Expose the port the app runs on
EXPOSE 8080

# Command to run the Go binary
CMD ["./backend-build"]