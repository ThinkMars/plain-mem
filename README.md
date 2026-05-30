# PlainMem

**PlainMem — plain, portable, long-term memory for AI agents.**

Install, run `pmem init`, done. Your agent remembers across sessions.

Inspired by the [GenericAgent](https://github.com/lsdefine/GenericAgent) memory architecture.

---

## Architecture

### Memory Layers

```
┌─────────────────────────────────────────────────────────────┐
│  L0: Constitution (Embedded in Source Code)                  │
│  • Immutable system rules                                   │
│  • Cannot be modified by any agent                          │
│  • Enforced by MemorySystem                                 │
└─────────────────────────────────────────────────────────────┘
                              ↓ enforced by
┌─────────────────────────────────────────────────────────────┐
│  L1: Insight Index (global_mem_insight.txt)                 │
│  • ≤30 lines, pointer-based routing                         │
│  • [RULES] section for behavioral guidelines                │
│  • L2/L3 section references                                 │
│  • Secret detection & line count validation                 │
└─────────────────────────────────────────────────────────────┘
                              ↓ points to
┌─────────────────────────────────────────────────────────────┐
│  L2: Fact Store (global_mem.txt)                            │
│  • Environment facts (paths, configs, APIs)                 │
│  • Organized by sections: [env], [config], [decision]       │
│  • Auto-sync with L1 index                                  │
└─────────────────────────────────────────────────────────────┘
                              ↓ references
┌─────────────────────────────────────────────────────────────┐
│  L3: Procedural Memory (pm/*.md, *.py)                      │
│  • Reusable SOPs and scripts                                │
│  • Index maintained in sop_index.json                       │
│  • Templates: Definition, Workflow, Red Lines                │
└─────────────────────────────────────────────────────────────┘
```

### Key Design Principles

1. **Layered Architecture**: Clear separation of concerns (L0 immutable rules → L3 reusable procedures)
2. **File-Based Storage**: Pure text files, no database dependencies, fully portable
3. **Index Synchronization**: L1 automatically updates when L2/L3 changes
4. **Immutable Constitution**: L0 lives in source code, cannot be modified by agents
5. **Validation First**: Verify before store, search before write
6. **MCP Compatibility**: Standard protocol for any AI host integration

---

## Quick Start

```bash
# Download the binary from Releases, or build from source:
git clone https://github.com/ThinkMars/plain-mem.git
cd plain-mem
go build -o ~/.local/bin/pmem ./cmd/pmem/

# Initialize — this writes the memory section into your AGENTS.md:
pmem init
```

That's it. Your AGENTS.md now tells your agent to use pmem.

---

## Commands

```bash
pmem read                    # Load memory into prompt (run at session start)
pmem agents                  # Print full usage instructions
pmem init                    # Write memory section into AGENTS.md

pmem add-fact <sec> <v>      # Add fact (appends if section exists)
pmem append-fact <sec> <v>   # Append line to existing section
pmem update-fact <sec> <v>   # Replace entire section (patch in place)

pmem add-rule <text>         # Add behavioral rule (L1)

pmem add-sop <name> [title]  # Create new SOP file (L3)
pmem update-sop <name> <v>   # Replace SOP content
pmem list-sops               # List all registered SOPs

pmem search <keyword>        # Search across L2 and L3
pmem consolidate             # Print consolidation instructions
pmem verify                  # Check memory integrity
pmem mcp                     # Run as MCP server over stdio
```

Use `-` instead of content to read from stdin (for multi-line input):

```bash
pmem update-fact env - <<'EOF'
python: 3.13
node: 22
EOF
```

---

## Examples

```bash
# Load memory
pmem read

# Store facts (auto-appends to existing sections)
pmem add-fact env "python: /opt/homebrew/bin/python3.12"
pmem add-fact config "port 8080, db sqlite"

# Add a rule
pmem add-rule "Always set timeouts on external API calls."

# Create and update SOPs
pmem add-sop deploy "Deploy Checklist"
pmem update-sop deploy - <<'EOF'
# Deploy Checklist
## Workflow
1. go test ./...
2. go build -o pmem ./cmd/pmem/
3. pmem verify
EOF
pmem list-sops

# Search
pmem search python

# End-of-session
pmem consolidate
pmem verify
```

---

## Storage

Memory is stored as plain files under `./pm/`:

```
pm/
├── global_mem_insight.txt      # L1 index
├── global_mem.txt              # L2 facts
├── sop_index.json              # L3 metadata
└── deploy_sop.md               # L3 SOP files
```

---

## Development

```bash
go build ./cmd/pmem/ && go vet ./... && go test ./internal/
```

---

## License

[MIT](https://github.com/ThinkMars/plain-mem/blob/main/LICENSE)
