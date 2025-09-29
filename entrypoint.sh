#!/bin/sh
set -e

# Download and install migrate if it doesn't exist
if [ ! -f /usr/local/bin/migrate ]; then
    echo "Installing migrate tool..."
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz
    mv migrate /usr/local/bin/migrate
    chmod +x /usr/local/bin/migrate
fi

# Run migrations
echo "Running database migrations..."
migrate -path /app/migrations -database "$DATABASE_URL" up

# Start the application
echo "Starting application..."
exec /app/pixshelf
