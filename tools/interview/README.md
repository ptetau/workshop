# Interview CLI

Interactive tool to turn PRDs and feature requests into a scaffold graph, then invoke the scaffold CLI to shape a compile-ready app.

## Usage

Interactively:

```powershell
go run ./tools/interview --root C:\temp\app --module workshop
```

Outputs:
- `.scaffold/interview/graph.json`
- `.scaffold/interview/graph.dot`

## Interview Rules

- Concepts: one word.
- Orchestrators and projections: short phrases (converted to symbols by removing spaces).
- Disambiguation questions: yes/no.
- Type questions: one word (e.g., `string`, `int`, `bool`, `time`).
- Stop any section by typing `done`.

## Notes

- The tool invokes `go run ./tools/scaffold init` with the generated flags.
- Use `--out` to change where the graph files are written.
