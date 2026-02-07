#!/bin/bash
# Database initialization script for Workshop Layer 1

DB_FILE="workshop.db"

echo "Initializing Workshop database..."

# Remove existing database
if [ -f "$DB_FILE" ]; then
    echo "Removing existing database..."
    rm "$DB_FILE"
fi

# Create database and apply migrations
echo "Creating database and applying migrations..."

# Apply all migrations in order
for migration in internal/adapters/storage/migrations/*.sql; do
    echo "Applying $(basename $migration)..."
    sqlite3 "$DB_FILE" < "$migration"
done

# Enable WAL mode for better concurrency
sqlite3 "$DB_FILE" "PRAGMA journal_mode=WAL;"

# Verify tables were created
echo ""
echo "Database tables:"
sqlite3 "$DB_FILE" ".tables"

echo ""
echo "Database initialized successfully at $DB_FILE"
