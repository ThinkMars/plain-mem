package internal

import (
	"fmt"
	"strings"
)

const (
	defaultInsight = `# [Global Memory Insight]
L2: (empty)
L3: (empty)

[RULES]
1. Search first: verify information before storing.
2. Only store data verified by successful actions.
3. Check memory layers in order: L1 → L2 → L3.
`
	secretPatterns = "api_key,password,secret,sk-,token:"
)

// Insight is the L1 pointer-based index (≤30 lines).
type Insight struct {
	storage  Storage
	path     string
	maxLines int
}

func NewInsight(storage Storage, path string, maxLines int) *Insight {
	i := &Insight{storage: storage, path: path, maxLines: maxLines}
	if !storage.Exists(path) {
		_ = storage.Write(path, defaultInsight)
	}
	return i
}

func (i *Insight) Content() string {
	s, err := i.storage.Read(i.path)
	if err != nil {
		return ""
	}
	return s
}

func (i *Insight) LineCount() int {
	return len(strings.Split(i.Content(), "\n"))
}

func (i *Insight) AddRule(rule string) error {
	current := i.GetConstitution()
	rules := trimEmpty(strings.Split(current, "\n"))
	rules = append(rules, rule)
	return i.SetConstitution(strings.Join(rules, "\n"))
}

func (i *Insight) GetConstitution() string {
	c := i.Content()
	idx := strings.Index(c, "[RULES]")
	if idx < 0 {
		return ""
	}
	return strings.TrimSpace(c[idx+len("[RULES]"):])
}

func (i *Insight) SetConstitution(text string) error {
	c := i.Content()
	idx := strings.Index(c, "[RULES]")
	if idx < 0 {
		c = strings.TrimSpace(c) + "\n\n[RULES]\n" + text
	} else {
		c = c[:idx+len("[RULES]")] + "\n" + text
	}
	return i.write(c)
}

func (i *Insight) UpdateLayerRef(layer int, value string) error {
	c := i.Content()
	prefix := fmt.Sprintf("L%d:", layer)
	var out []string
	updated := false
	for line := range SplitSeq(c, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), prefix) {
			out = append(out, prefix+" "+value)
			updated = true
		} else {
			out = append(out, line)
		}
	}
	if !updated {
		idx := 0
		for j, ln := range out {
			if strings.HasPrefix(strings.TrimSpace(ln), "[RULES]") {
				idx = j
				break
			}
		}
		if idx == 0 {
			idx = len(out)
		}
		before := out[:idx]
		after := out[idx:]
		out = append(append(before, prefix+" "+value), after...)
	}
	return i.write(strings.Join(out, "\n"))
}

func (i *Insight) Validate() []string {
	var issues []string
	if i.LineCount() > i.maxLines {
		issues = append(issues, fmt.Sprintf("L1 line count %d exceeds max %d", i.LineCount(), i.maxLines))
	}
	lower := strings.ToLower(i.Content())
	for _, pat := range strings.Split(secretPatterns, ",") {
		if strings.Contains(lower, pat) {
			issues = append(issues, fmt.Sprintf("Possible secret detected: '%s'", pat))
		}
	}
	return issues
}

func (i *Insight) ExportForPrompt() string {
	return strings.TrimSpace(i.Content())
}

func (i *Insight) write(content string) error {
	lines := strings.Split(content, "\n")
	if len(lines) > i.maxLines {
		excess := len(lines) - i.maxLines
		lines = lines[:i.maxLines]
		lines = append(lines, fmt.Sprintf("# WARNING: %d line(s) trimmed (max %d)", excess, i.maxLines))
		content = strings.Join(lines, "\n")
	}
	return i.storage.Write(i.path, content)
}

func trimEmpty(ss []string) []string {
	var out []string
	for _, s := range ss {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
