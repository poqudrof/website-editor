# Build stage
FROM golang:1.25.1-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY backend/go.mod backend/go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY backend/*.go ./

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o site-editor .

# Runtime stage with Node and Claude CLI
FROM node:20-alpine

# Install required dependencies
RUN apk add --no-cache \
    ca-certificates \
    curl \
    bash \
    git \
    vim \
    nano \
    sqlite-libs

# Install Claude CLI globally
RUN npm install -g @anthropic-ai/claude-code

# Create app and workspace directories with proper permissions
RUN mkdir -p /app /workspace && \
    chown -R node:node /app /workspace

# Set working directory
WORKDIR /app

# Copy the Go binary from builder and set permissions
COPY --from=builder /app/site-editor .
RUN chown node:node /app/site-editor && \
    chmod +x /app/site-editor

# Switch to node user
USER node

# Expose port
EXPOSE 9000

# Run the application
CMD ["./site-editor"]
