# Build stage
FROM golang:1.23.2-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY . .
RUN go build -ldflags="-s -w" -o main .

# Final stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/*.txt .
EXPOSE 8080
CMD ["./main"]