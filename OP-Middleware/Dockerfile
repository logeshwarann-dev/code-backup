# Build stage
FROM golang:1.23.2-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY . .
ENV GOFLAGS="-mod=vendor"
# Build the Go application
RUN go build -ldflags="-s -w" -o main .

# Final stage
FROM alpine:latest
WORKDIR /app
COPY . .
COPY --from=builder /app/main .
COPY --from=builder /app/*.txt .
EXPOSE 8090
CMD ["./main"]