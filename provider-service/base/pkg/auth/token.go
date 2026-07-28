package auth

import (
	"fmt"
	"os"
	"strings"
)

// TokenSource returns a bearer token to attach to outgoing requests.
type TokenSource interface {
	Token() (string, error)
}

// FileTokenSource reads a bearer token from a file on each call, so rotated
// projected service-account tokens are picked up.
type FileTokenSource struct {
	Path string
}

func NewFileTokenSource(path string) *FileTokenSource {
	return &FileTokenSource{Path: path}
}

func (s *FileTokenSource) Token() (string, error) {
	data, err := os.ReadFile(s.Path)
	if err != nil {
		return "", fmt.Errorf("reading token from %q: %w", s.Path, err)
	}
	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", fmt.Errorf("token file %q is empty", s.Path)
	}
	return token, nil
}
