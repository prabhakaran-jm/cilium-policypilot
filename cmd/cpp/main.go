package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/prabhakaran-jm/cilium-policypilot/internal/hubble"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "cpp",
		Short: "Cilium PolicyPilot CLI",
		Long:  "Learn from Hubble flows, propose minimal Cilium policies, verify them safely, and explain results.",
	}

	root.AddCommand(cmdLearn(), cmdPropose(), cmdVerify(), cmdExplain())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func cmdLearn() *cobra.Command {
	var inputFile string
	var captureDuration string

	cmd := &cobra.Command{
		Use:   "learn",
		Short: "Capture or read Hubble flows",
		Long:  "Read flows from a JSON file or capture them from Hubble CLI.\nIf no input file is provided, attempts to read from out/flows.json.",
		RunE: func(cmd *cobra.Command, args []string) error {
			outputFile := "out/flows.json"

			// Ensure output directory exists
			if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

			var collection *hubble.FlowCollection
			var err error

			// If input file is provided, read from it
			if inputFile != "" {
				fmt.Printf("Reading flows from %s...\n", inputFile)
				collection, err = hubble.ReadFlowsFromFile(inputFile)
				if err != nil {
					return fmt.Errorf("failed to read flows from file: %w", err)
				}
			} else {
				// Try to read from default location
				defaultFile := "out/flows.json"
				if _, err := os.Stat(defaultFile); err == nil {
					fmt.Printf("Reading flows from %s...\n", defaultFile)
					collection, err = hubble.ReadFlowsFromFile(defaultFile)
					if err != nil {
						return fmt.Errorf("failed to read flows from file: %w", err)
					}
				} else {
					// No existing file, create empty collection
					fmt.Println("No existing flows file found. Creating empty collection.")
					fmt.Println("Tip: Use 'hubble observe -o json > out/flows.json' to capture flows, or")
					fmt.Println("     provide an input file with --input flag.")
					collection = &hubble.FlowCollection{
						Schema: "cpp.flows.v1",
						Flows:  []*hubble.Flow{},
					}
				}
			}

			// Parse flows to validate and get statistics
			parsedFlows, err := hubble.ParseFlows(collection)
			if err != nil {
				return fmt.Errorf("failed to parse flows: %w", err)
			}

			fmt.Printf("Loaded %d flows (parsed %d successfully)\n", len(collection.Flows), len(parsedFlows))

			// Write to output file
			if err := hubble.WriteFlowsToFile(collection, outputFile); err != nil {
				return fmt.Errorf("failed to write flows: %w", err)
			}

			fmt.Printf("Flows saved to %s\n", outputFile)
			return nil
		},
	}

	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input flows JSON file (default: out/flows.json)")
	cmd.Flags().StringVarP(&captureDuration, "duration", "d", "", "Duration to capture flows (e.g., '--since 5m' or '--last 100')")

	return cmd
}

func cmdPropose() *cobra.Command {
	return &cobra.Command{
		Use:   "propose",
		Short: "Synthesize minimal Cilium policy",
		RunE: func(cmd *cobra.Command, args []string) error {
			sample := `apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: sample
spec:
  endpointSelector:
    matchLabels:
      app: demo
  ingress:
  - fromEndpoints:
    - matchLabels:
        app: demo
    toPorts:
    - ports:
      - port: "8080"
        protocol: TCP
`
			if err := os.MkdirAll("out", 0o755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}
			return os.WriteFile("out/policy.yaml", []byte(sample), 0o644)
		},
	}
}

func cmdVerify() *cobra.Command {
	return &cobra.Command{
		Use:   "verify",
		Short: "Run verification placeholder",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Verify: would spin kind + replay flows here")
			return nil
		},
	}
}

func cmdExplain() *cobra.Command {
	return &cobra.Command{
		Use:   "explain",
		Short: "Generate simple HTML report",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Explain: writing out/report.html")
			if err := os.MkdirAll("out", 0o755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}
			html := "# PolicyPilot Report\n\n```mermaid\ngraph TD; frontend-->catalog;\n```\n"
			return os.WriteFile("out/report.html", []byte(html), 0o644)
		},
	}
}
