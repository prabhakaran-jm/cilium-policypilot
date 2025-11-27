package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FilePath validates that a file path exists and is readable
func FilePath(path string) error {
	if path == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", path)
		}
		return fmt.Errorf("cannot access file %s: %w", path, err)
	}

	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", path)
	}

	return nil
}

// OutputPath validates and creates the directory for an output file path
func OutputPath(path string) error {
	if path == "" {
		return fmt.Errorf("output path cannot be empty")
	}

	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory %s: %w", dir, err)
		}
	}

	return nil
}

// Namespace validates a Kubernetes namespace name
func Namespace(ns string) error {
	if ns == "" {
		return nil // Empty namespace is valid (means all namespaces)
	}

	if len(ns) > 63 {
		return fmt.Errorf("namespace name too long (max 63 characters): %s", ns)
	}

	// Basic validation: alphanumeric and hyphens, must start/end with alphanumeric
	if !isValidK8sName(ns) {
		return fmt.Errorf("invalid namespace name: %s (must be lowercase alphanumeric with hyphens)", ns)
	}

	return nil
}

// isValidK8sName validates Kubernetes resource names
func isValidK8sName(name string) bool {
	if len(name) == 0 || len(name) > 253 {
		return false
	}

	// Must start and end with alphanumeric
	if !isAlphanumeric(rune(name[0])) || !isAlphanumeric(rune(name[len(name)-1])) {
		return false
	}

	// Can contain lowercase alphanumeric and hyphens
	for _, r := range name {
		if !isAlphanumeric(r) && r != '-' {
			return false
		}
	}

	return true
}

func isAlphanumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
}

// FileExtension validates that a file has the expected extension
func FileExtension(path string, expectedExt string) error {
	ext := strings.ToLower(filepath.Ext(path))
	expectedExt = strings.ToLower(expectedExt)

	if !strings.HasPrefix(expectedExt, ".") {
		expectedExt = "." + expectedExt
	}

	if ext != expectedExt {
		return fmt.Errorf("file must have %s extension, got %s", expectedExt, ext)
	}

	return nil
}
