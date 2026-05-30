package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ThinkMars/plain-mem/internal"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	arg := os.Args[1]
	if arg == "--help" || arg == "-h" {
		printUsage()
		os.Exit(0)
	}
	if arg == "--version" || arg == "-v" {
		fmt.Println("pmem version " + internal.Version)
		os.Exit(0)
	}

	mem, err := internal.NewMemorySystem(internal.DefaultMemoryDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "read":
		fmt.Println(mem.BuildSystemPrompt())

	case "agents":
		fmt.Print(internal.AgentsPrompt)

	case "init":
		changed, err := writeAgentsMd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if changed {
			fmt.Println("[pmem] AGENTS.md updated with memory section.")
		} else {
			fmt.Println("[pmem] AGENTS.md already up to date.")
		}

	case "add-fact":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "Usage: pmem add-fact <section> <content-or->")
			os.Exit(1)
		}
		section := os.Args[2]
		content := readContent(os.Args[3:])
		if err := mem.AddFact(section, content); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[pmem] Fact '%s' stored.\n", section)

	case "append-fact":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "Usage: pmem append-fact <section> <content-or->")
			os.Exit(1)
		}
		section := os.Args[2]
		content := readContent(os.Args[3:])
		if err := mem.AppendFact(section, content); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[pmem] Appended to '%s'.\n", section)

	case "update-fact":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "Usage: pmem update-fact <section> <content-or->")
			os.Exit(1)
		}
		section := os.Args[2]
		content := readContent(os.Args[3:])
		if err := mem.UpdateFact(section, content); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[pmem] Fact '%s' updated.\n", section)

	case "add-rule":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: pmem add-rule <text>")
			os.Exit(1)
		}
		rule := strings.Join(os.Args[2:], " ")
		if err := mem.AddRule(rule); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("[pmem] Rule added.")

	case "add-sop":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: pmem add-sop <name> [title]")
			os.Exit(1)
		}
		name := os.Args[2]
		title := ""
		if len(os.Args) > 3 && os.Args[3] != "-" {
			title = strings.Join(os.Args[3:], " ")
		}
		if title == "" {
			title = name
		}
		if err := mem.AddSop(name, title); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[pmem] SOP '%s' created.\n", name)

	case "update-sop":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "Usage: pmem update-sop <name> <content-or->")
			os.Exit(1)
		}
		name := os.Args[2]
		content := readContent(os.Args[3:])
		if err := mem.UpdateSop(name, content); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[pmem] SOP '%s' updated.\n", name)

	case "list-sops":
		entries := mem.ListSops()
		if len(entries) == 0 {
			fmt.Println("[pmem] No SOPs registered.")
			break
		}
		fmt.Println("[L3 SOPs]")
		for _, e := range entries {
			fmt.Printf("  %s  %s\n", e.Filename, e.Title)
		}

	case "search":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: pmem search <keyword>")
			os.Exit(1)
		}
		keyword := os.Args[2]
		results := mem.Search(keyword)
		if len(results) == 0 {
			fmt.Printf("[pmem] No results for '%s'\n", keyword)
			break
		}
		for _, r := range results {
			fmt.Printf("[L%d] %s: %s\n", r.Layer, r.Location, r.Snippet)
		}

	case "mcp":
		if err := internal.RunMCPServer(mem); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "consolidate":
		fmt.Println(mem.BuildConsolidationPrompt())

	case "verify":
		issues := mem.Verify()
		if len(issues) == 0 {
			fmt.Println("[pmem] All layers intact.")
		} else {
			for _, issue := range issues {
				fmt.Printf("[pmem] FAIL: %s\n", issue)
			}
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func readContent(args []string) string {
	if len(args) == 1 && args[0] == "-" {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading stdin: %v\n", err)
			os.Exit(1)
		}
		return strings.TrimSpace(string(b))
	}
	return strings.Join(args, " ")
}

func writeAgentsMd() (bool, error) {
	const agentsFile = "AGENTS.md"
	const marker = "## Memory\n"

	section := "## Memory\n\n" + internal.AgentsPrompt

	existing, err := os.ReadFile(agentsFile)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}

	content := string(existing)
	idx := strings.Index(content, marker)
	if idx >= 0 {
		end := strings.Index(content[idx+len(marker):], "\n## ")
		if end >= 0 {
			content = content[:idx] + section + "\n" + content[idx+len(marker)+end:]
		} else {
			content = content[:idx] + section + "\n"
		}
	} else {
		if len(content) > 0 && content[len(content)-1] != '\n' {
			content += "\n"
		}
		content += "\n" + section
	}

	changed := strings.TrimRight(content, "\n ") != strings.TrimRight(string(existing), "\n ")
	if !changed {
		return false, nil
	}
	return true, os.WriteFile(agentsFile, []byte(strings.TrimRight(content, "\n ")+"\n"), 0o644)
}

func printUsage() {
	fmt.Println(`PlainMem — layered long-term memory for AI agents

Read:
  pmem read                 Load memory context for prompt injection
  pmem agents               Print memory integration instructions
  pmem init                 Write memory section into AGENTS.md

L2 Facts:
  pmem add-fact <sec> <v>   Add fact (appends if section exists)
  pmem append-fact <sec> <v> Append line to existing section
  pmem update-fact <sec> <v> Replace entire section content

L1 Rules:
  pmem add-rule <text>      Add behavioral rule

L3 SOPs:
  pmem add-sop <name> [title] Create new SOP file (template)
  pmem update-sop <name> <v>  Replace SOP content
  pmem list-sops              List all registered SOPs

Other:
  pmem search <keyword>     Search across L2 and L3
  pmem consolidate          Print consolidation instructions
  pmem verify               Check memory integrity

MCP (stdio mode):
  pmem mcp                  Run as MCP server over stdio

Flags:
  --help, -h                Show this help
  --version, -v             Show version

Use '-' as content argument to read from stdin (for multi-line input).`)
}
