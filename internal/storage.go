package internal

import (
	"os"
	"path/filepath"
	"strings"
)

// FileStorage implements Storage using the local file system.
type FileStorage struct {
	baseDir string
}

func NewFileStorage(baseDir string) (*FileStorage, error) {
	abs, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, err
	}
	return &FileStorage{baseDir: abs}, nil
}

func (s *FileStorage) resolve(path string) (string, error) {
	clean := filepath.Clean(path)
	abs := filepath.Join(s.baseDir, clean)
	if !strings.HasPrefix(abs, s.baseDir) {
		return "", os.ErrPermission
	}
	return abs, nil
}

func (s *FileStorage) Read(path string) (string, error) {
	full, err := s.resolve(path)
	if err != nil {
		return "", err
	}
	b, err := os.ReadFile(full)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *FileStorage) ReadLines(path string) ([]string, error) {
	full, err := s.resolve(path)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(full)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(b), "\n")
	return lines, nil
}

func (s *FileStorage) Write(path string, content string) error {
	full, err := s.resolve(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	return os.WriteFile(full, []byte(content), 0o644)
}

func (s *FileStorage) Append(path string, content string) error {
	full, err := s.resolve(path)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(full, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

func (s *FileStorage) Patch(path string, oldContent string, newContent string) (bool, error) {
	full, err := s.resolve(path)
	if err != nil {
		return false, err
	}
	b, err := os.ReadFile(full)
	if err != nil {
		return false, err
	}
	content := string(b)
	if !strings.Contains(content, oldContent) {
		return false, nil
	}
	content = strings.Replace(content, oldContent, newContent, 1)
	return true, os.WriteFile(full, []byte(content), 0o644)
}

func (s *FileStorage) Exists(path string) bool {
	full, err := s.resolve(path)
	if err != nil {
		return false
	}
	_, err = os.Stat(full)
	return err == nil
}

func (s *FileStorage) Delete(path string) error {
	full, err := s.resolve(path)
	if err != nil {
		return err
	}
	return os.Remove(full)
}

func (s *FileStorage) DeleteDir(path string) error {
	full, err := s.resolve(path)
	if err != nil {
		return err
	}
	return os.RemoveAll(full)
}

func (s *FileStorage) ListDir(path string, pattern string) ([]string, error) {
	full, err := s.resolve(path)
	if err != nil {
		return nil, err
	}
	if pattern == "" {
		pattern = "*"
	}
	matches, err := filepath.Glob(filepath.Join(full, pattern))
	if err != nil {
		return nil, err
	}
	var rel []string
	for _, m := range matches {
		r, _ := filepath.Rel(s.baseDir, m)
		rel = append(rel, r)
	}
	return rel, nil
}

func (s *FileStorage) MakeDir(path string) error {
	full, err := s.resolve(path)
	if err != nil {
		return err
	}
	return os.MkdirAll(full, 0o755)
}

func (s *FileStorage) FileSize(path string) (int, error) {
	full, err := s.resolve(path)
	if err != nil {
		return 0, err
	}
	info, err := os.Stat(full)
	if err != nil {
		return 0, err
	}
	return int(info.Size()), nil
}
