package internal

// L0Constitution is embedded in source code — immutable by any agent.
const L0Constitution = `[L0: Memory Constitution — immutable system rules]

1. No Execution, No Memory — only store data verified by a successful tool call.
   Never store guesses, common knowledge, or unexecuted plans.

2. Sanctity of Verified Data — never delete or overwrite verified facts.
   When editing, use patch semantics (update in place, don't delete and recreate).

3. No Volatile State — never store timestamps, session IDs, PIDs, or
   system-specific absolute paths in L1-L3.

4. Minimum Sufficient Pointer — L1 index ≤ 30 lines (hard cap).
   Keep only the shortest identifier needed to locate details.
   Details belong in L2 facts or L3 SOPs.

5. Classification:
   - Environment-specific verified fact → L2 (global_mem.txt)
   - Universal behavioral rule / pitfall → L1 [RULES] (one sentence)
   - Reusable task workflow → L3 (create an SOP file under pm/)
   - Common knowledge or easy to reproduce → discard, do not store`

// ReservedFilenames are core files that cannot be created/deleted via SOP API.
var ReservedFilenames = map[string]bool{
	"global_mem_insight.txt": true,
	"global_mem.txt":         true,
	"sop_index.json":         true,
}

const systemPromptHeader = `Memory Layers:
  L1: global_mem_insight.txt — pointer-based index (≤30 lines)
  L2: global_mem.txt — environment-specific verified facts
  L3: pm/*.md, *.py — SOPs and utility scripts`

const Version = "0.2.0"
const DefaultSearchResultLimit = 30
const DefaultMemoryDir = "./pm"

const AgentsPrompt = `[Memory Integration]

Start every session with ` + "`pmem read`" + ` — it loads the constitution,
verified facts, SOPs, and behavioral rules.

**Write restriction: Only pmem may write to memory. Read directly from
./pm/ is OK, but all modifications must go through pmem commands.**

### Storage timing
- During task: store immediately after verifying a fact (command succeeds,
  path confirmed, decision reached).
- End of session: run ` + "`pmem consolidate`" + ` and follow its instructions.
- Never store guesses, common knowledge, or unverified information.

### Layer classification
  L1: cross-session rules / learned pitfalls → ` + "`pmem add-rule <text>`" + `
  L2: verified project facts (paths, configs, APIs, decisions)
      → ` + "`pmem add-fact <section> <content>`" + ` (appends if section exists)
      → ` + "`pmem append-fact <section> <content>`" + `
      → ` + "`pmem update-fact <section> <content>`" + ` (patch in place)
  L3: reusable SOPs (workflows with 3+ steps)
      → ` + "`pmem add-sop <name> [title]`" + ` (creates template)
      → ` + "`pmem update-sop <name> <content>`" + `
      → ` + "`pmem list-sops`" + `

### Section naming for L2
Use short tags: ` + "`[env]` `[decision]` `[pitfall]` `[config]` `[api]` `[architecture]` `[workflow]`" + `

### Multi-line content
Use heredoc with stdin: ` + "`pmem update-fact env - <<'EOF' ... EOF`" + `

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
`
