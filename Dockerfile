FROM golang:1.23-alpine AS builder

# Set the working directory
WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o pixshelf ./cmd/server/main.go

# Create a minimal image
FROM alpine:3.19

WORKDIR /app

# Install dependencies and migrate tool directly
RUN apk add --no-cache ca-certificates tzdata curl && \
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

# Copy the binary from the builder stage
COPY --from=builder /app/pixshelf /app/pixshelf

# Copy migrations, templates, and static files
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/templates /app/templates
COPY --from=builder /app/static /app/static

# Create directory for image storage
RUN mkdir -p /app/static/images && chmod -R 755 /app/static/images

# Create startup script properly with actual newlines
RUN printf '#!/bin/sh\nset -e\necho "Running database migrations..."\nmigrate -path /app/migrations -database "$DATABASE_URL" up\necho "Starting application..."\nexec /app/pixshelf\n' > /app/start.sh && \
    chmod +x /app/start.sh

# Set the entrypoint
ENTRYPOINT ["/app/start.sh"]

# Expose the port
EXPOSE 8080
