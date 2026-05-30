
## Memory

[Memory Integration]

Start every session with `pmem read` — it loads the constitution,
verified facts, SOPs, and behavioral rules.

**Write restriction: Only pmem may write to memory. Read directly from
./pm/ is OK, but all modifications must go through pmem commands.**

### Storage timing
- During task: store immediately after verifying a fact (command succeeds,
  path confirmed, decision reached).
- End of session: run `pmem consolidate` and follow its instructions.
- Never store guesses, common knowledge, or unverified information.

### Layer classification
  L1: cross-session rules / learned pitfalls → `pmem add-rule <text>`
  L2: verified project facts (paths, configs, APIs, decisions)
      → `pmem add-fact <section> <content>` (appends if section exists)
      → `pmem append-fact <section> <content>`
      → `pmem update-fact <section> <content>` (patch in place)
  L3: reusable SOPs (workflows with 3+ steps)
      → `pmem add-sop <name> [title]` (creates template)
      → `pmem update-sop <name> <content>`
      → `pmem list-sops`

### Section naming for L2
Use short tags: `[env]` `[decision]` `[pitfall]` `[config]` `[api]` `[architecture]` `[workflow]`

### Multi-line content
Use heredoc with stdin: `pmem update-fact env - <<'EOF' ... EOF`

### Commands reference
  pmem read             Load memory context
  pmem add-fact <s> <c> Add or append to L2 section
  pmem append-fact <s> <c> Append line to L2 section
  pmem update-fact <s> <c> Replace L2 section (patch in place)
  pmem add-rule <text>    Add L1 behavioral rule
  pmem add-sop <n> [t]    Create L3 SOP file
  pmem update-sop <n> <c> Replace L3 SOP content
  pmem list-sops          List all L3 SOPs
  pmem search <keyword>   Search across L2 and L3
  pmem consolidate        Print consolidation instructions
  pmem verify             Check memory integrity

## Git Commit Message Convention

Format: `<type>(<scope>): <subject>`

Type: feat/fix/docs/style/refactor/perf/test/chore/revert

Rules: subject concisely describes the change, scope is optional. Body (if any) explains "what and why", not "how". Do not describe unnecessary details. After generating the commit message, ask the user to review if it is reasonable. Must use English in subject and body!.
