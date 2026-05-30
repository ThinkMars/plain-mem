package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// MemorySystem is the top-level orchestrator.
type MemorySystem struct {
	mu      sync.Mutex
	Config  MemoryConfig
	Storage Storage
	Insight *Insight
	Facts   *FactStore
	Sops    *Sops
}

func NewMemorySystem(baseDir string) (*MemorySystem, error) {
	cfg := DefaultConfig(baseDir)
	storage, err := NewFileStorage(cfg.BaseDir)
	if err != nil {
		return nil, err
	}

	m := &MemorySystem{
		Config:  cfg,
		Storage: storage,
	}

	m.Insight = NewInsight(storage, cfg.InsightFilename, cfg.MaxInsightLines)
	m.Facts = NewFactStore(storage, cfg.FactFilename)
	m.Sops = NewSops(storage, cfg.BaseDir, cfg.SopMetadataFile)

	return m, nil
}

func (m *MemorySystem) BuildSystemPrompt() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.buildSystemPrompt()
}

func (m *MemorySystem) buildSystemPrompt() string {
	var b strings.Builder

	b.WriteString(L0Constitution)
	b.WriteString("\n\n")
	b.WriteString(systemPromptHeader)
	b.WriteString("\n\n[Memory] (../pm)")

	if insight := m.Insight.ExportForPrompt(); insight != "" {
		b.WriteString("\n\n../pm/global_mem_insight.txt:\n")
		b.WriteString(insight)
	}

	if facts := m.Facts.ExportForPrompt(true); facts != "" && !strings.Contains(strings.Split(facts, "\n")[0], "empty") {
		b.WriteString("\n\n")
		b.WriteString(facts)
	}

	if sops := m.Sops.ExportForPrompt(); sops != "" {
		b.WriteString("\n\n")
		b.WriteString(sops)
	}

	return b.String()
}

func (m *MemorySystem) Search(keyword string) []SearchResult {
	m.mu.Lock()
	defer m.mu.Unlock()

	var results []SearchResult
	results = append(results, m.Facts.Search(keyword)...)
	results = append(results, m.Sops.Search(keyword)...)

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	return results
}

func (m *MemorySystem) Verify() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	var issues []string

	if !m.Storage.Exists(m.Config.InsightFilename) {
		issues = append(issues, "L1 insight file missing")
	} else {
		issues = append(issues, m.Insight.Validate()...)
	}

	if !m.Storage.Exists(m.Config.FactFilename) {
		issues = append(issues, "L2 fact file missing")
	}

	if len(m.Sops.ListAll()) == 0 {
		issues = append(issues, "L3 no SOPs registered")
	}

	expected := m.Insight.Content()
	for _, section := range m.Facts.ListSections() {
		if !strings.Contains(expected, section) {
			issues = append(issues, fmt.Sprintf("L2 section '%s' not referenced in L1 index", section))
		}
	}

	return issues
}

func (m *MemorySystem) BuildConsolidationPrompt() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	prompt := `[MEMORY CONSOLIDATION] Review the completed task and distill verified, long-term-valuable information into persistent memory.

Classification:
  - Environment-specific verified facts (paths, configs, APIs) → L2
  - Reusable task workflows / hard-earned techniques → L3 (SOP)
  - Universal behavioral rules / learned pitfalls → L1 [RULES] (one sentence)

Rules:
  1. Only store data verified by successful actions (No Execution, No Memory).
  2. Never overwrite existing memory — use patch operations.
  3. When adding to L2/L3, update L1 index pointers.
  4. Skip anything you could easily reproduce or that's common knowledge.
`
	return prompt + "\n" + m.buildSystemPrompt()
}

// Path returns the resolved baseDir for this system.
func (m *MemorySystem) Path() string {
	abs, _ := filepath.Abs(m.Config.BaseDir)
	return abs
}

// MemDir returns the path relative to the storage backend root.
func (m *MemorySystem) MemDir() string {
	abs, _ := filepath.Abs(m.Config.BaseDir)
	cwd, _ := os.Getwd()
	rel, _ := filepath.Rel(cwd, abs)
	if rel == "." || rel == "" {
		return "./pm"
	}
	return rel
}

// AddFact adds or appends an L2 fact section and syncs the L1 index.
func (m *MemorySystem) AddFact(section, content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.Facts.AddSection(section, content); err != nil {
		return err
	}
	m.syncFactInsight()
	return nil
}

// AppendFact appends a line to an existing L2 fact section and syncs the L1 index.
func (m *MemorySystem) AppendFact(section, content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing := m.Facts.GetSection(section)
	if existing == nil {
		if err := m.Facts.AddSection(section, content); err != nil {
			return err
		}
	} else {
		if err := m.Facts.UpdateSection(section, existing.Content+"\n"+content); err != nil {
			return err
		}
	}
	m.syncFactInsight()
	return nil
}

// UpdateFact replaces an L2 fact section content and syncs the L1 index.
func (m *MemorySystem) UpdateFact(section, content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.Facts.UpdateSection(section, content); err != nil {
		return err
	}
	m.syncFactInsight()
	return nil
}

// AddRule adds a behavioral rule to the L1 insight index.
func (m *MemorySystem) AddRule(rule string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Insight.AddRule(rule)
}

// AddSop creates a new L3 SOP file and syncs the L1 index.
func (m *MemorySystem) AddSop(name, title string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.Sops.Create(name, title, ""); err != nil {
		return err
	}
	m.syncSopInsight()
	return nil
}

// UpdateSop replaces an L3 SOP file content and syncs the L1 index.
func (m *MemorySystem) UpdateSop(name, content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.Sops.Update(name, content); err != nil {
		return err
	}
	m.syncSopInsight()
	return nil
}

// ListSops returns all registered L3 SOP entries.
func (m *MemorySystem) ListSops() []SOPEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Sops.ListAll()
}

func (m *MemorySystem) syncFactInsight() {
	sections := m.Facts.ListSections()
	value := "(empty)"
	if len(sections) > 0 {
		value = strings.Join(sections, " | ")
	}
	_ = m.Insight.UpdateLayerRef(2, value)
}

func (m *MemorySystem) syncSopInsight() {
	var filenames []string
	for _, e := range m.Sops.ListAll() {
		filenames = append(filenames, e.Filename)
	}
	value := "(empty)"
	if len(filenames) > 0 {
		value = strings.Join(filenames, " | ")
	}
	_ = m.Insight.UpdateLayerRef(3, value)
}
