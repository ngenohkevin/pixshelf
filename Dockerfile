FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install required packages
RUN apk add --no-cache git

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pixshelf ./cmd/server

# Final stage
FROM alpine:latest

# Set working directory
WORKDIR /app

# Install necessary runtime dependencies
RUN apk --no-cache add ca-certificates

# Copy the built binary from the builder stage
COPY --from=builder /app/pixshelf .

# Copy static files and templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates

# Create a non-root user to run the application
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN chown -R appuser:appgroup /app
USER appuser

# Expose port
EXPOSE 8080

# Set environment variables
ENV ENV=production
ENV PORT=8080
ENV DATABASE_URL=postgres://postgres:postgres@db:5432/pixshelf?sslmode=disable
ENV IMAGE_STORAGE=/app/static/images
ENV BASE_URL=http://localhost:8080

# Run the application
CMD ["./pixshelf"]
