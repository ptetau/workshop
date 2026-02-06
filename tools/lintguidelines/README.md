# Guidelines Linter

Validate deterministic parts of the project guidelines. The linter checks naming, coupling, routing, and storage isolation rules based on the repo layout.

Run from the repository root.

## Usage

```powershell
go run ./tools/lintguidelines --format text
```

Strict mode (non-zero exit if violations exist):

```powershell
go run ./tools/lintguidelines --strict
```

JSON output:

```powershell
go run ./tools/lintguidelines --format json
```

## Rules Enforced

1) Naming (deterministic subset)
- Flags identifiers containing banned abbreviations: `usr`, `amt`, `cfg`, `tmp`, `svc`
- Allowed abbreviations: `ID`, `URL`, `HTTP`, `JSON`, `SQL`, `UUID`, `API`

2) Concept coupling
- Concept packages (`internal/domain/<concept>`) must not import other concepts, `internal/application`, or `internal/adapters`.

3) Route discipline
- GET handlers must call `projections.Query*` and must not call `orchestrators.Execute*`.
- POST/PUT/DELETE handlers must call `orchestrators.Execute*` and must not call `projections.Query*`.

4) Storage isolation
- Storage packages (`internal/adapters/storage/<concept>`) must have a matching concept directory.
- Storage package name must match the folder name.
- Storage packages must not import other concepts.

## Notes

- Generated routes (`internal/adapters/http/routes.go`) are checked by the route rule only.
- The linter uses folder conventions as the source of truth.
