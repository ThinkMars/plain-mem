package internal

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

func RunMCPServer(sys *MemorySystem) error {
	server := newMCPServer(sys)
	return server.Run(context.Background(), &mcp.StdioTransport{})
}

func newMCPServer(sys *MemorySystem) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "pmem",
		Version: Version,
	}, nil)

	type emptyInput struct{}

	mcp.AddTool[emptyInput, any](server, &mcp.Tool{
		Name:        "read_memory",
		Description: "Load complete memory context (L0 constitution, L1 insight, L2 facts, L3 SOPs) for prompt injection",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
		return textResult(sys.BuildSystemPrompt()), nil, nil
	})

	type addFactInput struct {
		Section string `json:"section" jsonschema:"Section name, e.g. env, config, api, decision, pitfall"`
		Content string `json:"content" jsonschema:"Fact content to store"`
	}

	mcp.AddTool[addFactInput, any](server, &mcp.Tool{
		Name:        "add_fact",
		Description: "Add a new L2 fact section or append content if the section already exists",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input addFactInput) (*mcp.CallToolResult, any, error) {
		if err := sys.AddFact(input.Section, input.Content); err != nil {
			return textResult(fmt.Sprintf("error: %v", err)), nil, nil
		}
		return textResult(fmt.Sprintf("Fact '%s' stored.", input.Section)), nil, nil
	})

	type appendFactInput struct {
		Section string `json:"section" jsonschema:"Existing section name to append to"`
		Content string `json:"content" jsonschema:"Content line to append"`
	}

	mcp.AddTool[appendFactInput, any](server, &mcp.Tool{
		Name:        "append_fact",
		Description: "Append a line to an existing L2 fact section",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input appendFactInput) (*mcp.CallToolResult, any, error) {
		if err := sys.AppendFact(input.Section, input.Content); err != nil {
			return textResult(fmt.Sprintf("error: %v", err)), nil, nil
		}
		return textResult(fmt.Sprintf("Appended to '%s'.", input.Section)), nil, nil
	})

	type updateFactInput struct {
		Section string `json:"section" jsonschema:"Section name to replace"`
		Content string `json:"content" jsonschema:"New content for the entire section"`
	}

	mcp.AddTool[updateFactInput, any](server, &mcp.Tool{
		Name:        "update_fact",
		Description: "Replace an entire L2 fact section (patch in place)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input updateFactInput) (*mcp.CallToolResult, any, error) {
		if err := sys.UpdateFact(input.Section, input.Content); err != nil {
			return textResult(fmt.Sprintf("error: %v", err)), nil, nil
		}
		return textResult(fmt.Sprintf("Fact '%s' updated.", input.Section)), nil, nil
	})

	type addRuleInput struct {
		Rule string `json:"rule" jsonschema:"Behavioral rule text to add (one sentence)"`
	}

	mcp.AddTool[addRuleInput, any](server, &mcp.Tool{
		Name:        "add_rule",
		Description: "Add a behavioral rule to L1 insight index under [RULES]",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input addRuleInput) (*mcp.CallToolResult, any, error) {
		if err := sys.AddRule(input.Rule); err != nil {
			return textResult(fmt.Sprintf("error: %v", err)), nil, nil
		}
		return textResult("Rule added."), nil, nil
	})

	type searchInput struct {
		Keyword string `json:"keyword" jsonschema:"Keyword to search across L2 facts and L3 SOPs"`
	}

	mcp.AddTool[searchInput, any](server, &mcp.Tool{
		Name:        "search_memory",
		Description: "Search across L2 facts and L3 SOPs for a keyword, returns sorted results with scores",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input searchInput) (*mcp.CallToolResult, any, error) {
		results := sys.Search(input.Keyword)
		if len(results) == 0 {
			return textResult(fmt.Sprintf("No results for '%s'", input.Keyword)), nil, nil
		}
		var b strings.Builder
		for _, r := range results {
			b.WriteString(fmt.Sprintf("[L%d] %s: %s\n", r.Layer, r.Location, r.Snippet))
		}
		return textResult(strings.TrimSpace(b.String())), nil, nil
	})

	type addSopInput struct {
		Name  string `json:"name" jsonschema:"SOP filename (with or without .md/.py extension)"`
		Title string `json:"title,omitempty" jsonschema:"Optional display title (defaults to name)"`
	}

	mcp.AddTool[addSopInput, any](server, &mcp.Tool{
		Name:        "add_sop",
		Description: "Create a new L3 SOP file with a template",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input addSopInput) (*mcp.CallToolResult, any, error) {
		title := input.Title
		if title == "" {
			title = input.Name
		}
		if err := sys.AddSop(input.Name, title); err != nil {
			return textResult(fmt.Sprintf("error: %v", err)), nil, nil
		}
		return textResult(fmt.Sprintf("SOP '%s' created.", input.Name)), nil, nil
	})

	type updateSopInput struct {
		Name    string `json:"name" jsonschema:"SOP filename to update"`
		Content string `json:"content" jsonschema:"Full new content for the SOP file (read from stdin supports multi-line)"`
	}

	mcp.AddTool[updateSopInput, any](server, &mcp.Tool{
		Name:        "update_sop",
		Description: "Replace an existing L3 SOP file's content",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input updateSopInput) (*mcp.CallToolResult, any, error) {
		if err := sys.UpdateSop(input.Name, input.Content); err != nil {
			return textResult(fmt.Sprintf("error: %v", err)), nil, nil
		}
		return textResult(fmt.Sprintf("SOP '%s' updated.", input.Name)), nil, nil
	})

	mcp.AddTool[emptyInput, any](server, &mcp.Tool{
		Name:        "list_sops",
		Description: "List all registered L3 SOP files with titles",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
		entries := sys.ListSops()
		if len(entries) == 0 {
			return textResult("No SOPs registered."), nil, nil
		}
		var b strings.Builder
		b.WriteString("[L3 SOPs]\n")
		for _, e := range entries {
			b.WriteString(fmt.Sprintf("  %s  %s\n", e.Filename, e.Title))
		}
		return textResult(strings.TrimSpace(b.String())), nil, nil
	})

	mcp.AddTool[emptyInput, any](server, &mcp.Tool{
		Name:        "verify",
		Description: "Check memory integrity across all layers (L1 line count, secret detection, L2/L3 cross-references)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
		issues := sys.Verify()
		if len(issues) == 0 {
			return textResult("All layers intact."), nil, nil
		}
		var b strings.Builder
		for _, issue := range issues {
			b.WriteString(fmt.Sprintf("FAIL: %s\n", issue))
		}
		return textResult(strings.TrimSpace(b.String())), nil, nil
	})

	mcp.AddTool[emptyInput, any](server, &mcp.Tool{
		Name:        "consolidate",
		Description: "Print consolidation instructions for end-of-session memory review",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
		return textResult(sys.BuildConsolidationPrompt()), nil, nil
	})

	return server
}
