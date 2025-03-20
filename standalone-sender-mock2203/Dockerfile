# Build Stage
FROM redhat/ubi9:latest AS builder

# Install necessary dependencies including Go 1.23.2 and OpenSSL
RUN yum install -y \
    gcc \
    make \
    && yum clean all

# Install Go 1.23.2 manually
ENV GO_VERSION=1.23.2
RUN curl -sSL https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz -o go.tar.gz && \
    tar -C /usr/local -xzf go.tar.gz && \
    rm go.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}"

# Verify Go installation
RUN go version

# Set working directory
WORKDIR /app

# Copy project files
COPY go.mod go.sum ./
COPY . .

ENV GOFLAGS="-mod=vendor"

# Build the Go application
RUN go build -ldflags="-s -w" -o main .

# Final Stage
FROM redhat/ubi9:latest

# Set working directory
WORKDIR /app

# Copy the executable and dependencies
COPY --from=builder /app /app

# Expose ports for your service
EXPOSE 8080

# Command to run the application
CMD ["./main"]
