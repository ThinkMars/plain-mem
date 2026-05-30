package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func tempDir(t *testing.T) string {
	t.Helper()
	d, err := os.MkdirTemp("", "agentic_memory_test_")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(d) })
	return d
}

func TestFileStorageBasic(t *testing.T) {
	d := tempDir(t)
	s, err := NewFileStorage(d)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.Write("test.txt", "hello world"); err != nil {
		t.Fatal(err)
	}
	if !s.Exists("test.txt") {
		t.Error("file should exist")
	}

	content, err := s.Read("test.txt")
	if err != nil {
		t.Fatal(err)
	}
	if content != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", content)
	}

	if err := s.Append("test.txt", " appendix"); err != nil {
		t.Fatal(err)
	}
	content, _ = s.Read("test.txt")
	if content != "hello world appendix" {
		t.Errorf("expected 'hello world appendix', got '%s'", content)
	}

	ok, err := s.Patch("test.txt", "hello", "HELLO")
	if err != nil || !ok {
		t.Fatal("patch failed")
	}
	content, _ = s.Read("test.txt")
	if content != "HELLO world appendix" {
		t.Errorf("expected 'HELLO world appendix', got '%s'", content)
	}

	if err := s.Delete("test.txt"); err != nil {
		t.Fatal(err)
	}
	if s.Exists("test.txt") {
		t.Error("file should be deleted")
	}
}

func TestFileStoragePathTraversal(t *testing.T) {
	d := tempDir(t)
	s, err := NewFileStorage(d)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.Read("../etc/passwd")
	if err == nil {
		t.Error("path traversal should be prevented")
	}
}

func TestInsightAddRule(t *testing.T) {
	d := tempDir(t)
	s, _ := NewFileStorage(d)
	insight := NewInsight(s, "insight.txt", 30)

	if err := insight.AddRule("Test rule one."); err != nil {
		t.Fatal(err)
	}
	c := insight.GetConstitution()
	if c != "1. Search first: verify information before storing.\n2. Only store data verified by successful actions.\n3. Check memory layers in order: L1 → L2 → L3.\nTest rule one." {
		t.Errorf("unexpected constitution: %s", c)
	}

	issues := insight.Validate()
	if len(issues) > 0 {
		t.Errorf("unexpected validation issues: %v", issues)
	}
}

func TestInsightLineLimit(t *testing.T) {
	d := tempDir(t)
	s, _ := NewFileStorage(d)
	insight := NewInsight(s, "insight.txt", 10)

	for i := 0; i < 15; i++ {
		_ = insight.AddRule("Rule")
	}
	if insight.LineCount() > 11 { // 10 + warning line
		t.Errorf("expected ≤11 lines, got %d", insight.LineCount())
	}
}

func TestFactStoreCRUD(t *testing.T) {
	d := tempDir(t)
	s, _ := NewFileStorage(d)
	facts := NewFactStore(s, filepath.Join(d, "facts.txt"))

	if err := facts.AddSection("paths", "python: /usr/bin/python3"); err != nil {
		t.Fatal(err)
	}
	sec := facts.GetSection("paths")
	if sec == nil {
		t.Fatal("section should exist")
	}
	if !contains(sec.Content, "python: /usr/bin/python3") {
		t.Errorf("unexpected content: %s", sec.Content)
	}

	if err := facts.UpdateSection("paths", "python: /usr/bin/python3.12"); err != nil {
		t.Fatal(err)
	}
	sec = facts.GetSection("paths")
	if !contains(sec.Content, "python3.12") {
		t.Error("update didn't persist")
	}

	if _, err := facts.RemoveSection("paths"); err != nil {
		t.Fatal(err)
	}
	if facts.GetSection("paths") != nil {
		t.Error("section should be removed")
	}
}

func TestFactStoreSearch(t *testing.T) {
	d := tempDir(t)
	s, _ := NewFileStorage(d)
	facts := NewFactStore(s, filepath.Join(d, "facts.txt"))

	_ = facts.AddSection("env", "python: 3.12\nshell: zsh")
	results := facts.Search("python")
	if len(results) == 0 {
		t.Error("search should find python")
	}
}

func TestSopsCreateAndSearch(t *testing.T) {
	d := tempDir(t)
	s, _ := NewFileStorage(d)
	sops := NewSops(s, d, filepath.Join(d, "sop_index.json"))

	err := sops.Create("test_sop.md", "Test SOP", "## Workflow\n1. Step one\n2. Step two")
	if err != nil {
		t.Fatal(err)
	}

	content, err := sops.Get("test_sop.md")
	if err != nil {
		t.Fatal(err)
	}
	if !contains(content, "Step one") {
		t.Error("content should contain 'Step one'")
	}

	results := sops.Search("Step")
	if len(results) == 0 {
		t.Error("search should find 'Step'")
	}
}

func TestSopsRejectsReservedFilenames(t *testing.T) {
	d := tempDir(t)
	s, _ := NewFileStorage(d)
	sops := NewSops(s, d, filepath.Join(d, "sop_index.json"))

	if err := sops.Create("global_mem.txt", "Bad", ""); err == nil {
		t.Error("should reject reserved filename")
	}
}

func TestMemorySystemBuildPrompt(t *testing.T) {
	d := tempDir(t)
	mem, err := NewMemorySystem(d)
	if err != nil {
		t.Fatal(err)
	}

	prompt := mem.BuildSystemPrompt()
	if !contains(prompt, "Memory Constitution") {
		t.Error("prompt should contain L0 constitution")
	}
	if !contains(prompt, "No Execution, No Memory") {
		t.Error("prompt should contain L0 rules")
	}
}

func TestMemorySystemSearchAcrossLayers(t *testing.T) {
	d := tempDir(t)
	mem, err := NewMemorySystem(d)
	if err != nil {
		t.Fatal(err)
	}

	_ = mem.Facts.AddSection("env", "python: 3.12")
	_ = mem.Sops.Create("py_sop.md", "Python Guide", "Use Python 3.12 for projects")

	results := mem.Search("python")
	if len(results) < 2 {
		t.Errorf("expected at least 2 results, got %d", len(results))
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
