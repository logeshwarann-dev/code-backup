# Use a slim base image to reduce image size
FROM golang:1.23.4-alpine AS builder

# Set working directory
WORKDIR /app

# Copy Go modules
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

# Use a smaller runtime image
FROM alpine:latest

# Set working directory
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/main .

# Expose the port
EXPOSE 8082

# Run the application
CMD ["./main"]