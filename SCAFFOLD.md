# Scaffold Tool Guide

## Overview
The `scaffold` tool generates a complete, production-ready Go web application following Clean Architecture principles. It supports **smart updates** that preserve your customizations when you re-run the tool with new fields or routes.

## Quick Start

### Initialize a New Project
```bash
cd myproject
go run ../tools/scaffold init \
  --module myapp \
  --concept User \
  --field User:Name:string \
  --field User:Email:string \
  --orchestrator RegisterUser \
  --param RegisterUser:Email:string \
  --param RegisterUser:Password:string \
  --route "POST:/users:RegisterUser"
```

### Add Fields to Existing Concept (Smart Update)
```bash
go run ../tools/scaffold init \
  --module myapp \
  --concept User \
  --field User:Name:string \
  --field User:Email:string \
  --field User:Age:int  # NEW FIELD
```

**Result**: The tool will:
- ✅ Add `Age int` to the `User` struct
- ✅ Preserve any custom methods you added to `User`
- ✅ Generate an `ALTER TABLE` migration
- ✅ Update the storage layer to handle the new field

## Architecture

### Generated Structure
```
myproject/
├── cmd/server/main.go              # HTTP server entry point
├── internal/
│   ├── domain/                     # Business logic (concepts)
│   │   └── user/
│   │       └── model.go            # User struct + methods
│   ├── application/                # Use cases
│   │   ├── orchestrators/          # Commands (write operations)
│   │   │   └── register_user.go
│   │   └── projections/            # Queries (read operations)
│   └── adapters/                   # External interfaces
│       ├── http/                   # Web layer
│       │   ├── routes.go           # Generated handlers
│       │   ├── web.go              # Middleware setup
│       │   ├── middleware/         # CSRF, security headers
│       │   └── templates/          # HTML views/forms
│       └── storage/                # Persistence
│           ├── migrations/         # SQL migrations
│           └── user/
│               ├── store.go        # Interface
│               └── sqlite_store.go # Implementation
└── .scaffold/
    └── state.json                  # Scaffold metadata
```

## Smart Updates

### How It Works

#### Go Code (AST-Based)
The tool uses Go's `go/parser` and `go/ast` packages to:
1. Parse existing `.go` files
2. Locate the target struct (e.g., `User`)
3. Add missing fields to the struct
4. Preserve all other code (methods, comments, imports)
5. Write back formatted code using `go/format`

**Example**:
```go
// Before (your custom code)
type User struct {
    ID   string
    Name string
}

func (u *User) SayHello() string { return "Hello, " + u.Name }

// After running scaffold with --field User:Age:int
type User struct {
    ID   string
    Name string
    Age  int  // ADDED
}

func (u *User) SayHello() string { return "Hello, " + u.Name } // PRESERVED
```

#### HTML Templates (String-Based)
For HTML forms and views, the tool:
1. Reads the existing HTML file
2. Searches for standard markers (`</form>`, `</dl>`)
3. Checks if the field already exists (by `name="Field"`)
4. Injects new HTML before the marker if missing

**Example**:
```html
<!-- Before -->
<form method="POST">
    <input type="hidden" name="gorilla.csrf.Token" value="...">
    <div class="form-group">
        <label for="id_Email">Email</label>
        <input type="text" id="id_Email" name="Email" required>
    </div>
    <button type="submit">Submit</button>
</form>

<!-- After adding --param RegisterUser:Password:string -->
<form method="POST">
    <input type="hidden" name="gorilla.csrf.Token" value="...">
    <div class="form-group">
        <label for="id_Email">Email</label>
        <input type="text" id="id_Email" name="Email" required>
    </div>
    <!-- INJECTED -->
    <div class="form-group">
        <label for="id_Password">Password</label>
        <input type="text" id="id_Password" name="Password" required>
    </div>
    <button type="submit">Submit</button>
</form>
```

### Limitations
- **Go**: Only adds fields to structs. Does not modify function bodies or other types.
- **HTML**: Assumes standard scaffold-generated structure. Custom layouts may require manual updates.
- **Renaming**: Not supported. Add new field, migrate data manually, then remove old field.

## Security

### CSRF Protection
All `POST`/`PUT`/`DELETE` routes are protected by `gorilla/csrf`:
- Token automatically injected into forms
- Validated on submission
- **Production**: Load CSRF key from environment variable (see `internal/adapters/http/web.go`)

### Security Headers
Applied by default via `middleware.SecurityHeaders`:
- `Content-Security-Policy`: Restricts script/style sources
- `X-Frame-Options: DENY`: Prevents clickjacking
- `X-Content-Type-Options: nosniff`: Prevents MIME sniffing
- `Referrer-Policy`: Controls referrer information

### Linter Enforcement
Run `go run ./tools/lintguidelines --strict` to verify:
- ✅ Security middleware is applied in `NewMux`
- ✅ Concepts don't import other concepts (isolation)
- ✅ GET handlers only call projections (CQRS)
- ✅ POST handlers only call orchestrators (CQRS)

## Common Workflows

### Add a New Route
```bash
go run ../tools/scaffold init \
  --module myapp \
  --route "GET:/users/{id}:GetUser"
```
- Regenerates `routes.go` with new handler
- Creates `get_user.html` template (if projection exists)

### Add a Method to a Concept
1. Run scaffold normally (no changes needed)
2. Edit `internal/domain/user/model.go`:
   ```go
   func (u *User) IsAdult() bool {
       return u.Age >= 18
   }
   ```
3. Re-run scaffold with new fields → method is preserved

### Migrate Database
```bash
# Migrations are auto-generated in internal/adapters/storage/migrations/
# Apply them manually or via your migration tool
sqlite3 app.db < internal/adapters/storage/migrations/001_create_user.sql
sqlite3 app.db < internal/adapters/storage/migrations/002_alter_user.sql
```

## Troubleshooting

### "struct X not found in source"
- Ensure you're using the exact concept name (case-sensitive)
- Check that `model.go` exists and is valid Go code

### "could not find closing </form> tag"
- HTML file may have custom structure
- Manually add the field or regenerate the file with `--force`

### Tests Fail After Update
- Run `go mod tidy` to ensure dependencies are current
- Check that custom code doesn't conflict with new fields
- Review migration SQL for correctness

## Advanced

### Force Regeneration
```bash
# WARNING: Overwrites existing files
go run ../tools/scaffold init --force --concept User ...
```

### Custom Types
Supported field types:
- `string`, `int`, `bool`, `time.Time`
- Custom types require manual editing after generation

### Testing
```bash
cd tools/scaffold
go test -v .  # Run all scaffold tests
```

## See Also
- [DB_GUIDE.md](../../DB_GUIDE.md) - Database patterns
- [PRD.md](../../PRD.md) - Product requirements
- `tools/lintguidelines` - Architecture linter
