FROM golang:1.25-alpine AS builder

# Install required dependencies
RUN apk add --no-cache protobuf protobuf-dev git bash

# Install protoc plugins with specific versions compatible with Go 1.21
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
RUN go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.26.3
WORKDIR /app

# Copy go.mod and go.sum files first for better caching
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# # Enable static linking for Go binaries
ENV CGO_ENABLED=0 GOOS=linux

# Build the server and gateway binaries without launching the local dev stack.
RUN mkdir -p bin \
	&& go build -v -o bin/server ./cmd/server \
	&& go build -v -o bin/toolplane-gateway ./cmd/proxy

# Create a smaller final image
FROM alpine:latest

WORKDIR /app

# Copy the built binaries from the builder stage
COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/bin/toolplane-gateway /app/toolplane-gateway

# Copy the entrypoint script
COPY entrypoint.sh /app/entrypoint.sh
RUN sed -i 's/\r$//' /app/entrypoint.sh && chmod +x /app/entrypoint.sh

# Expose the ports
EXPOSE 9001
EXPOSE 8080
EXPOSE 9102

ENV TOOLPLANE_DATABASE_URL=postgres://user:password@localhost:5432/user?sslmode=disable
ENV TOOLPLANE_METRICS_LISTEN=:9102
# Run the entrypoint script
CMD ["/app/entrypoint.sh"]
