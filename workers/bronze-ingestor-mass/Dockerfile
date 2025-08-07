# Stage 1: Builder
FROM golang:1.24 AS builder

# Set working directory
WORKDIR /src

# Copy the entire project (including go.mod at root)
COPY . .

# Build static binary for Linux
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Build the target binary from the root module
RUN go build -o /bronze-ingestor-mass ./workers/bronze-ingestor-mass

# Stage 2: Runtime
FROM alpine:latest

# Create working directory in the final image
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /bronze-ingestor-mass .

# Set entrypoint
ENTRYPOINT ["./bronze-ingestor-mass"]
