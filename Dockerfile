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

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite-libs

# Create app directory
WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/site-editor .

# Expose port
EXPOSE 9000

# Run the application
CMD ["./site-editor"]
