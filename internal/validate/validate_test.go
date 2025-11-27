package validate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFilePath(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "non-existent file",
			path:    "/nonexistent/file.txt",
			wantErr: true,
		},
		{
			name:    "existing file",
			path:    tmpFile.Name(),
			wantErr: false,
		},
		{
			name:    "directory instead of file",
			path:    os.TempDir(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FilePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("FilePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOutputPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "valid path",
			path:    "out/test.json",
			wantErr: false,
		},
		{
			name:    "nested path",
			path:    "out/subdir/test.json",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up after test
			defer func() {
				if dir := filepath.Dir(tt.path); dir != "." && dir != "" {
					os.RemoveAll(dir)
				}
			}()

			err := OutputPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("OutputPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNamespace(t *testing.T) {
	tests := []struct {
		name    string
		ns      string
		wantErr bool
	}{
		{
			name:    "empty namespace",
			ns:      "",
			wantErr: false, // Empty is valid (means all namespaces)
		},
		{
			name:    "valid namespace",
			ns:      "default",
			wantErr: false,
		},
		{
			name:    "valid namespace with hyphens",
			ns:      "kube-system",
			wantErr: false,
		},
		{
			name:    "namespace too long",
			ns:      "a" + string(make([]byte, 64)),
			wantErr: true,
		},
		{
			name:    "namespace with uppercase",
			ns:      "Default",
			wantErr: true,
		},
		{
			name:    "namespace starting with hyphen",
			ns:      "-invalid",
			wantErr: true,
		},
		{
			name:    "namespace ending with hyphen",
			ns:      "invalid-",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Namespace(tt.ns)
			if (err != nil) != tt.wantErr {
				t.Errorf("Namespace() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileExtension(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectedExt string
		wantErr     bool
	}{
		{
			name:        "correct extension",
			path:        "test.json",
			expectedExt: ".json",
			wantErr:     false,
		},
		{
			name:        "wrong extension",
			path:        "test.txt",
			expectedExt: ".json",
			wantErr:     true,
		},
		{
			name:        "extension without dot",
			path:        "test.json",
			expectedExt: "json",
			wantErr:     false,
		},
		{
			name:        "case insensitive",
			path:        "test.JSON",
			expectedExt: ".json",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FileExtension(tt.path, tt.expectedExt)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileExtension() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
