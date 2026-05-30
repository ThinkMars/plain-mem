package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const sopTemplate = `# %s

## Definition / Core Principles


## When to Use / Trigger Conditions


## Prerequisites


## Workflow


## Red Lines / Critical Constraints


## Common Mistakes

`

// Sops is L3: procedural memory — SOP files and utility scripts.
type Sops struct {
	storage   Storage
	baseDir   string
	indexPath string
	index     map[string]SOPEntry
}

func NewSops(storage Storage, baseDir string, indexPath string) *Sops {
	s := &Sops{
		storage:   storage,
		baseDir:   baseDir,
		indexPath: indexPath,
	}
	s.loadIndex()
	return s
}

func (s *Sops) loadIndex() {
	data, err := s.storage.Read(s.indexPath)
	if err != nil {
		s.index = make(map[string]SOPEntry)
		s.rebuildIndex()
		return
	}
	if err := json.Unmarshal([]byte(data), &s.index); err != nil {
		s.index = make(map[string]SOPEntry)
		s.rebuildIndex()
	}
}

func (s *Sops) saveIndex() error {
	return s.storage.Write(s.indexPath, jsonPretty(s.index))
}

func (s *Sops) rebuildIndex() {
	s.index = make(map[string]SOPEntry)
	for _, e := range s.scanFiles() {
		s.index[e.Filename] = e
	}
	_ = s.saveIndex()
}

func (s *Sops) scanFiles() []SOPEntry {
	var entries []SOPEntry
	files, _ := s.storage.ListDir("", "")
	for _, f := range files {
		ext := filepath.Ext(f)
		if ext != ".md" && ext != ".py" {
			continue
		}
		fname := filepath.Base(f)
		if ReservedFilenames[fname] {
			continue
		}
		title := strings.TrimSuffix(fname, ext)
		title = strings.TrimSuffix(title, "_sop")
		size, _ := s.storage.FileSize(f)
		content, err := s.storage.Read(f)
		if err == nil && ext == ".md" {
			for line := range SplitSeq(content, "\n") {
				if strings.HasPrefix(line, "# ") {
					title = strings.TrimSpace(line[2:])
					break
				}
			}
		}
		if title == "" {
			title = strings.TrimSuffix(fname, ext)
			title = strings.TrimSuffix(title, "_sop")
		}
		createdAt := ""
		if entry, ok := s.index[fname]; ok {
			createdAt = entry.CreatedAt
		}
		entries = append(entries, SOPEntry{
			Filename:  fname,
			Title:     title,
			SizeBytes: size,
			CreatedAt: createdAt,
		})
	}
	return entries
}

func (s *Sops) ListAll() []SOPEntry {
	s.loadIndex()
	var entries []SOPEntry
	for _, v := range s.index {
		entries = append(entries, v)
	}
	return entries
}

func (s *Sops) Get(filename string) (string, error) {
	return s.storage.Read(filename)
}

func (s *Sops) Create(filename string, title string, content string) error {
	if ReservedFilenames[filename] {
		return fmt.Errorf("'%s' is a reserved core file and cannot be created as an SOP", filename)
	}
	if !strings.HasSuffix(filename, ".md") && !strings.HasSuffix(filename, ".py") {
		filename += ".md"
	}

	if content == "" {
		content = fmt.Sprintf(sopTemplate, title)
	}

	now := time.Now().Format("2006-01-02 15:04")
	if err := s.storage.Write(filename, content); err != nil {
		return err
	}

	entry, exists := s.index[filename]
	if exists {
		entry.UpdatedAt = now
		s.index[filename] = entry
	} else {
		s.index[filename] = SOPEntry{
			Filename:  filename,
			Title:     title,
			SizeBytes: len(content),
			CreatedAt: now,
			UpdatedAt: now,
		}
	}
	_ = s.saveIndex()
	return nil
}

func (s *Sops) Update(filename string, content string) error {
	if !strings.HasSuffix(filename, ".md") && !strings.HasSuffix(filename, ".py") {
		filename += ".md"
	}
	if !s.storage.Exists(filename) {
		return fmt.Errorf("SOP '%s' does not exist", filename)
	}
	now := time.Now().Format("2006-01-02 15:04")
	if err := s.storage.Write(filename, content); err != nil {
		return err
	}
	title := ""
	for line := range SplitSeq(content, "\n") {
		if strings.HasPrefix(line, "# ") {
			title = strings.TrimSpace(line[2:])
			break
		}
	}
	if entry, ok := s.index[filename]; ok {
		entry.UpdatedAt = now
		entry.SizeBytes = len(content)
		if title != "" {
			entry.Title = title
		}
		s.index[filename] = entry
	} else {
		s.index[filename] = SOPEntry{
			Filename:  filename,
			Title:     title,
			SizeBytes: len(content),
			CreatedAt: now,
			UpdatedAt: now,
		}
	}
	_ = s.saveIndex()
	return nil
}

func (s *Sops) Delete(filename string) error {
	if ReservedFilenames[filename] {
		return fmt.Errorf("'%s' is a reserved core file and cannot be deleted", filename)
	}
	if !s.storage.Exists(filename) {
		return os.ErrNotExist
	}
	if err := s.storage.Delete(filename); err != nil {
		return err
	}
	delete(s.index, filename)
	_ = s.saveIndex()
	return nil
}

func (s *Sops) Search(keyword string) []SearchResult {
	kw := strings.ToLower(keyword)
	var results []SearchResult
	for _, entry := range s.ListAll() {
		content, err := s.storage.Read(entry.Filename)
		if err != nil {
			continue
		}
		nameMatch := strings.Contains(strings.ToLower(entry.Title), kw)
		lineNum := 0
		for line := range SplitSeq(content, "\n") {
			lineNum++
			if strings.Contains(strings.ToLower(line), kw) {
				score := 0.5
				if nameMatch {
					score = 1.0
				}
				snippet := line
				if len(snippet) > 200 {
					snippet = snippet[:200]
				}
				results = append(results, SearchResult{
					Layer:    3,
					Location: fmt.Sprintf("L3 %s:%d", entry.Filename, lineNum),
					Snippet:  snippet,
					Score:    score,
				})
			}
		}
		if len(results) >= DefaultSearchResultLimit {
			break
		}
	}
	return results
}

func (s *Sops) ExportForPrompt() string {
	entries := s.ListAll()
	if len(entries) == 0 {
		return ""
	}
	var lines []string
	lines = append(lines, "[Procedural Memory - L3]")
	for _, e := range entries {
		lines = append(lines, fmt.Sprintf("  %s: %s", e.Filename, e.Title))
	}
	return strings.Join(lines, "\n")
}
