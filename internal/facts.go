package internal

import (
	"fmt"
	"regexp"
	"strings"
)

var sectionPattern = regexp.MustCompile(`(?m)^## \[(.+)\]`)
var multiNewlinePattern = regexp.MustCompile(`\n{3,}`)

// FactStore is L2: verified environment facts organized by sections.
type FactStore struct {
	storage Storage
	path    string
}

func NewFactStore(storage Storage, path string) *FactStore {
	fs := &FactStore{storage: storage, path: path}
	if !storage.Exists(path) {
		_ = storage.Write(path, "# [Global Memory - L2]\n")
	}
	return fs
}

func (f *FactStore) Content() string {
	s, err := f.storage.Read(f.path)
	if err != nil {
		return "# [Global Memory - L2]\n"
	}
	return s
}

func (f *FactStore) ListSections() []string {
	var names []string
	for _, m := range sectionPattern.FindAllStringSubmatch(f.Content(), -1) {
		names = append(names, m[1])
	}
	return names
}

func (f *FactStore) GetSection(name string) *FactSection {
	for _, s := range f.parseSections() {
		if s.Name == name {
			return &s
		}
	}
	return nil
}

func (f *FactStore) AddSection(name string, content string) error {
	existing := f.GetSection(name)
	if existing != nil {
		// Append content to existing section
		oldBlock := fmt.Sprintf("## [%s]\n%s", name, existing.Content)
		newBlock := fmt.Sprintf("## [%s]\n%s\n%s", name, existing.Content, strings.TrimSpace(content))
		c := f.Content()
		c = strings.Replace(c, oldBlock, newBlock, 1)
		if err := f.storage.Write(f.path, c); err != nil {
			return err
		}
		return nil
	}
	current := strings.TrimSpace(f.Content())
	section := fmt.Sprintf("\n\n## [%s]\n%s", name, strings.TrimSpace(content))
	if err := f.storage.Write(f.path, current+section); err != nil {
		return err
	}
	return nil
}

func (f *FactStore) UpdateSection(name string, content string) error {
	existing := f.GetSection(name)
	if existing == nil {
		return f.AddSection(name, content)
	}

	c := f.Content()
	oldBlock := fmt.Sprintf("## [%s]\n%s", name, existing.Content)
	newBlock := fmt.Sprintf("## [%s]\n%s", name, strings.TrimSpace(content))

	if strings.Contains(c, oldBlock) {
		c = strings.Replace(c, oldBlock, newBlock, 1)
	} else {
		header := fmt.Sprintf("## [%s]", name)
		var out []string
		inSection := false
		for line := range SplitSeq(c, "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == header {
				out = append(out, line)
				out = append(out, strings.TrimSpace(content))
				inSection = true
			} else if inSection && strings.HasPrefix(trimmed, "## [") {
				out = append(out, line)
				inSection = false
			} else if !inSection {
				out = append(out, line)
			}
		}
		c = strings.Join(out, "\n")
	}
	return f.storage.Write(f.path, c)
}

func (f *FactStore) RemoveSection(name string) (bool, error) {
	if f.GetSection(name) == nil {
		return false, nil
	}
	header := fmt.Sprintf("## [%s]", name)
	var out []string
	skip := false
	for line := range SplitSeq(f.Content(), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == header {
			skip = true
			continue
		}
		if skip && strings.HasPrefix(trimmed, "## [") {
			skip = false
		}
		if !skip {
			out = append(out, line)
		}
	}
	result := strings.TrimSpace(strings.Join(out, "\n"))
	result = multiNewlinePattern.ReplaceAllString(result, "\n\n")
	if err := f.storage.Write(f.path, result+"\n"); err != nil {
		return false, err
	}
	return true, nil
}

func (f *FactStore) Search(keyword string) []SearchResult {
	var results []SearchResult
	kw := strings.ToLower(keyword)
	for _, s := range f.parseSections() {
		for line := range SplitSeq(s.Content, "\n") {
			if strings.Contains(strings.ToLower(line), kw) {
				score := 0.5
				if strings.Contains(strings.ToLower(s.Name), kw) {
					score = 1.0
				}
				results = append(results, SearchResult{
					Layer: 2, Location: "L2# " + s.Name,
					Snippet: line, Score: score,
				})
			}
		}
	}
	return results
}

func (f *FactStore) ExportForPrompt(compact bool) string {
	sections := f.parseSections()
	if len(sections) == 0 {
		return "L2: (empty)"
	}
	var lines []string
	lines = append(lines, "[Global Facts - L2]")
	for _, s := range sections {
		if compact {
			firstLine := strings.Split(strings.TrimSpace(s.Content), "\n")[0]
			if len(firstLine) > 120 {
				firstLine = firstLine[:120]
			}
			lines = append(lines, fmt.Sprintf("  [%s] %s", s.Name, firstLine))
		} else {
			lines = append(lines, fmt.Sprintf("\n## [%s]", s.Name))
			lines = append(lines, strings.TrimSpace(s.Content))
		}
	}
	return strings.Join(lines, "\n")
}

func (f *FactStore) parseSections() []FactSection {
	var sections []FactSection
	var currentName string
	var currentLines []string
	for line := range SplitSeq(f.Content(), "\n") {
		m := sectionPattern.FindStringSubmatch(line)
		if m != nil {
			if currentName != "" {
				sections = append(sections, FactSection{
					Name:    currentName,
					Content: strings.TrimRight(strings.Join(currentLines, "\n"), "\n "),
				})
			}
			currentName = m[1]
			currentLines = nil
		} else if currentName != "" {
			currentLines = append(currentLines, line)
		}
	}
	if currentName != "" {
		sections = append(sections, FactSection{
			Name:    currentName,
			Content: strings.TrimRight(strings.Join(currentLines, "\n"), "\n "),
		})
	}
	return sections
}
