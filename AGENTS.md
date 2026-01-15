# Agent Instructions

See **CLAUDE.md** for complete agent context and instructions.

This file exists for compatibility with tools that look for AGENTS.md.

> **Recovery**: Run `gt prime` after compaction, clear, or new session

Full context is injected by `gt prime` at session start.

## Code Quality Rules

### 5x Iteration Rule
When working on code, **iterate and improve it 5 times before committing**:
1. Write the initial implementation
2. Review for bugs, edge cases, error handling
3. Review for clarity, naming, structure
4. Review for performance, efficiency
5. Final polish - comments, formatting, simplification

Only commit after completing all 5 iterations. This ensures higher quality code with fewer review cycles.

### Test Fixing Rule
**Always fix broken tests before completing your task.** When you:
- Break existing tests with your changes → fix them
- Discover pre-existing test failures → fix them
- Add new functionality → add tests for it

Never leave tests in a failing state. Run `go test ./...` before marking work complete.

### Context Refresh Rule
**If idle and under 15% context remaining, refresh your session.**

When you:
- Complete a task and have no pending work
- Notice context is below 15%

Run `gt crew refresh <your-name>` to get a fresh context. This prevents working with degraded context that leads to lower quality output.
