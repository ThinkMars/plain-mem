package internal

import "encoding/json"

// SearchResult represents a single search hit.
type SearchResult struct {
	Layer    int     `json:"layer"`
	Location string  `json:"location"`
	Snippet  string  `json:"snippet"`
	Score    float64 `json:"score"`
}

// MemoryConfig holds configuration for the memory system.
type MemoryConfig struct {
	BaseDir         string
	InsightFilename string
	FactFilename    string
	SopMetadataFile string
	MaxInsightLines int
	SearchLimit     int
}

func DefaultConfig(baseDir string) MemoryConfig {
	return MemoryConfig{
		BaseDir:         baseDir,
		InsightFilename: "global_mem_insight.txt",
		FactFilename:    "global_mem.txt",
		SopMetadataFile: "sop_index.json",
		MaxInsightLines: 30,
		SearchLimit:     DefaultSearchResultLimit,
	}
}

// FactSection represents a single L2 fact section.
type FactSection struct {
	Name    string
	Content string
}

// SOPEntry holds metadata about a stored SOP file.
type SOPEntry struct {
	Filename  string `json:"filename"`
	Title     string `json:"title"`
	Category  string `json:"category,omitempty"`
	SizeBytes int    `json:"size_bytes"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Storage defines the storage backend interface.
type Storage interface {
	Read(path string) (string, error)
	ReadLines(path string) ([]string, error)
	Write(path string, content string) error
	Append(path string, content string) error
	Patch(path string, oldContent string, newContent string) (bool, error)
	Exists(path string) bool
	Delete(path string) error
	DeleteDir(path string) error
	ListDir(path string, pattern string) ([]string, error)
	MakeDir(path string) error
	FileSize(path string) (int, error)
}

func jsonPretty(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}
