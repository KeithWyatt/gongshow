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
