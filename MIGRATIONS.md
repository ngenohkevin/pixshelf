# Database Migrations Guide

PixShelf uses [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations.

## Creating a New Migration

To create a new migration, use the following command:

```bash
make migrate-create name=migration_name
```

This will create two new files in the `migrations` directory:
- `000001_migration_name.up.sql` - Contains SQL to apply the migration
- `000001_migration_name.down.sql` - Contains SQL to revert the migration

## Running Migrations

To apply all pending migrations:

```bash
make migrate-up
```

To rollback the most recent migration:

```bash
make migrate-down
```

## Migration Files Structure

Each migration consists of two files:

1. **Up Migration**: Contains SQL statements to apply changes to the database. For example:

```sql
-- 000001_create_images_table.up.sql
CREATE TABLE IF NOT EXISTS images (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    file_path VARCHAR(255) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    size_bytes BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_images_name ON images (name);
CREATE INDEX IF NOT EXISTS idx_images_created_at ON images (created_at);
```

2. **Down Migration**: Contains SQL statements to revert the changes. For example:

```sql
-- 000001_create_images_table.down.sql
DROP TABLE IF EXISTS images;
```

## Migration Best Practices

1. **Keep migrations small and focused**
   - Each migration should do one thing well
   - Easier to debug and rollback if needed

2. **Always create a down migration**
   - Ensures you can rollback changes if needed

3. **Test migrations before applying to production**
   - Run both up and down migrations in a test environment

4. **Use transactions where appropriate**
   - Some migrations (like those involving multiple tables) should be wrapped in transactions
