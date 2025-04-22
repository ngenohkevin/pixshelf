# PixShelf

PixShelf is a simple image hosting application built with Go, HTMX, Alpine.js, and Tailwind CSS.

## Features

- Upload images with metadata
- View gallery of uploaded images
- View image details
- Edit image metadata
- Delete images
- Search images by name or description
- Dark mode UI
- Responsive design

## Tech Stack

- **Backend:** Go, Gin, pgx, sqlc
- **Database:** PostgreSQL
- **Frontend:** HTMX, Alpine.js, Templ, Tailwind CSS
- **Deployment:** Coolify

## Development Setup

### Prerequisites

- Go 1.21 or higher
- PostgreSQL
- sqlc (for generating database code)
- Templ (for compiling templates)

### Installation

2. Install dependencies:

```bash
go mod download
```

3. Set up the database:

```bash
# Create a PostgreSQL database named 'pixshelf'
createdb pixshelf

# Run migrations
go run internal/db/migrate/main.go up
```

4. Generate sqlc code:

```bash
sqlc generate
```

5. Compile Templ templates:

```bash
templ generate
```

6. Run the application:

```bash
go run cmd/server/main.go
```

The application will be available at http://localhost:8080.

## Environment Variables

The application uses the following environment variables:

- `PORT`: Server port (default: 8080)
- `DATABASE_URL`: PostgreSQL connection string
- `ENV`: Environment name (default: "development")
- `IMAGE_STORAGE`: Path to store images (default: "./static/images")
- `BASE_URL`: Base URL for generating image URLs (default: "http://localhost:8080")

## License

MIT

## Author

Kevin Ngenoh - [ngenohkevin](https://github.com/ngenohkevin)
