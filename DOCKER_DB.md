# Docker PostgreSQL Database Setup

This document describes how to work with the PostgreSQL database running in Docker for PixShelf.

## Getting Started

The project is configured to use PostgreSQL in Docker. Here are the essential commands for working with it:

### Starting the Database

To start only the PostgreSQL database container:

```bash
make db-start
```

This will start the PostgreSQL container with the following configuration:
- Host: localhost
- Port: 5432
- Username: postgres
- Password: postgres
- Database: pixshelf

### Working with Migrations

After the database is running, you can apply migrations:

```bash
make migrate-up
```

To rollback the latest migration:

```bash
make migrate-down
```

To create a new migration:

```bash
make migrate-create name=your_migration_name
```

### Generating SQLc Code

After applying migrations, you can generate SQLc code:

```bash
make sqlc
```

### Full Docker Setup

To start both the database and the application:

```bash
make docker-up
```

To stop all containers:

```bash
make docker-down
```

To view logs:

```bash
make docker-logs
```

## Database Connection Details

When developing locally and connecting to the Docker database:

- Connection String: `postgres://postgres:postgres@localhost:5432/pixshelf?sslmode=disable`

When running the application inside Docker:

- Connection String: `postgres://postgres:postgres@db:5432/pixshelf?sslmode=disable`

## Data Persistence

The database data is stored in a Docker volume called `postgres_data`, ensuring your data persists between container restarts.
