# Database initialization script for Workshop Layer 1

$DB_FILE = "workshop.db"

Write-Host "Initializing Workshop database..."

# Remove existing database
if (Test-Path $DB_FILE) {
    Write-Host "Removing existing database..."
    Remove-Item $DB_FILE
}

# Create database and apply migrations
Write-Host "Creating database and applying migrations..."

# Apply all migrations in order
Get-ChildItem "internal\adapters\storage\migrations\*.sql" | Sort-Object Name | ForEach-Object {
    Write-Host "Applying $($_.Name)..."
    Get-Content $_.FullName | sqlite3 $DB_FILE
}

# Enable WAL mode for better concurrency
"PRAGMA journal_mode=WAL;" | sqlite3 $DB_FILE

# Verify tables were created
Write-Host ""
Write-Host "Database tables:"
".tables" | sqlite3 $DB_FILE

Write-Host ""
Write-Host "Database initialized successfully at $DB_FILE"
