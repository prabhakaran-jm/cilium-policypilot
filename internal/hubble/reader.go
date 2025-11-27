package hubble

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// HubbleReader handles reading flows from Hubble
type HubbleReader struct {
	// Path to Hubble CLI (default: "hubble")
	HubbleCLI string

	// Output directory for flow files
	OutputDir string
}

// NewHubbleReader creates a new HubbleReader with default settings
func NewHubbleReader() *HubbleReader {
	return &HubbleReader{
		HubbleCLI: "hubble",
		OutputDir: "out",
	}
}

// CaptureFlows captures flows from Hubble CLI and saves to file
// This runs: hubble observe -o json > output_file
func (r *HubbleReader) CaptureFlows(duration string, outputFile string) error {
	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build hubble observe command
	args := []string{"observe", "-o", "json"}

	// Add duration if specified (e.g., "--since 5m" or "--last 100")
	if duration != "" {
		args = append(args, duration)
	}

	// Execute hubble observe command
	cmd := exec.Command(r.HubbleCLI, args...)

	// Capture output to file
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	cmd.Stdout = outFile
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute hubble observe: %w", err)
	}

	return nil
}

// ReadFlowsFromHubbleAPI reads flows directly from Hubble API
// This is a placeholder for future API integration
func (r *HubbleReader) ReadFlowsFromHubbleAPI(endpoint string) (*FlowCollection, error) {
	// TODO: Implement Hubble API client
	return nil, fmt.Errorf("hubble API integration not yet implemented")
}
