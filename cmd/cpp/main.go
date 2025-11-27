package main

import (
	"fmt"
	"os"

	"github.com/prabhakaran-jm/cilium-policypilot/internal/hubble"
	"github.com/prabhakaran-jm/cilium-policypilot/internal/synth"
	"github.com/prabhakaran-jm/cilium-policypilot/internal/validate"
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
	var outputFile string
	var captureDuration string
	var hubbleEndpoint string

	cmd := &cobra.Command{
		Use:   "learn",
		Short: "Capture or read Hubble flows",
		Long:  "Read flows from a JSON file or capture them from Hubble CLI.\nIf no input file is provided, attempts to read from out/flows.json.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set default output file if not provided
			if outputFile == "" {
				outputFile = "out/flows.json"
			}

			// Validate output path
			if err := validate.OutputPath(outputFile); err != nil {
				return fmt.Errorf("invalid output path: %w", err)
			}

			// Validate output file extension
			if err := validate.FileExtension(outputFile, ".json"); err != nil {
				return fmt.Errorf("output file must be JSON: %w", err)
			}

			var collection *hubble.FlowCollection
			var err error

			// If input file is provided, validate and read from it
			if inputFile != "" {
				if err := validate.FilePath(inputFile); err != nil {
					return fmt.Errorf("invalid input file: %w", err)
				}
				if err := validate.FileExtension(inputFile, ".json"); err != nil {
					return fmt.Errorf("input file must be JSON: %w", err)
				}
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

			// Validate collection schema
			if collection.Schema == "" {
				return fmt.Errorf("invalid flows file: missing schema field")
			}

			// Parse flows to validate and get statistics
			parsedFlows, err := hubble.ParseFlows(collection)
			if err != nil {
				return fmt.Errorf("failed to parse flows: %w", err)
			}

			fmt.Printf("Loaded %d flows (parsed %d successfully)\n", len(collection.Flows), len(parsedFlows))

			if len(collection.Flows) > 0 && len(parsedFlows) == 0 {
				fmt.Fprintf(os.Stderr, "Warning: No flows could be parsed. Check that flows have required fields (source, destination, l4).\n")
			}

			// Write to output file
			if err := hubble.WriteFlowsToFile(collection, outputFile); err != nil {
				return fmt.Errorf("failed to write flows: %w", err)
			}

			fmt.Printf("Flows saved to %s\n", outputFile)
			return nil
		},
	}

	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input flows JSON file (default: out/flows.json)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output flows JSON file (default: out/flows.json)")
	cmd.Flags().StringVarP(&captureDuration, "duration", "d", "", "Duration to capture flows (e.g., '--since 5m' or '--last 100')")
	cmd.Flags().StringVar(&hubbleEndpoint, "hubble-endpoint", "", "Hubble API endpoint (for future API integration)")

	return cmd
}

func cmdPropose() *cobra.Command {
	var inputFile string
	var outputFile string
	var namespaceFilter string

	cmd := &cobra.Command{
		Use:   "propose",
		Short: "Synthesize minimal Cilium policy",
		Long:  "Generate CiliumNetworkPolicies from parsed flows.\nReads flows from out/flows.json (or specified input file) and generates policies.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set default input file if not provided
			if inputFile == "" {
				inputFile = "out/flows.json"
			}

			// Set default output file if not provided
			if outputFile == "" {
				outputFile = "out/policy.yaml"
			}

			// Validate input file
			if err := validate.FilePath(inputFile); err != nil {
				return fmt.Errorf("invalid input file: %w", err)
			}
			if err := validate.FileExtension(inputFile, ".json"); err != nil {
				return fmt.Errorf("input file must be JSON: %w", err)
			}

			// Validate output path
			if err := validate.OutputPath(outputFile); err != nil {
				return fmt.Errorf("invalid output path: %w", err)
			}
			if err := validate.FileExtension(outputFile, ".yaml"); err != nil {
				// Also accept .yml extension
				if err2 := validate.FileExtension(outputFile, ".yml"); err2 != nil {
					return fmt.Errorf("output file must be YAML (.yaml or .yml): %w", err)
				}
			}

			// Validate namespace filter if provided
			if namespaceFilter != "" {
				if err := validate.Namespace(namespaceFilter); err != nil {
					return fmt.Errorf("invalid namespace filter: %w", err)
				}
			}

			// Read flows
			fmt.Printf("Reading flows from %s...\n", inputFile)
			collection, err := hubble.ReadFlowsFromFile(inputFile)
			if err != nil {
				return fmt.Errorf("failed to read flows: %w", err)
			}

			// Validate collection
			if collection == nil {
				return fmt.Errorf("invalid flows file: collection is nil")
			}
			if collection.Schema == "" {
				return fmt.Errorf("invalid flows file: missing schema field")
			}

			// Parse flows
			parsedFlows, err := hubble.ParseFlows(collection)
			if err != nil {
				return fmt.Errorf("failed to parse flows: %w", err)
			}

			if len(parsedFlows) == 0 {
				return fmt.Errorf("no valid flows found to generate policies from")
			}

			// Apply namespace filter if provided
			if namespaceFilter != "" {
				filtered := make([]*hubble.ParsedFlow, 0)
				for _, flow := range parsedFlows {
					// Include flows where source or destination matches the namespace
					if flow.SourceNamespace == namespaceFilter || flow.DestNamespace == namespaceFilter {
						filtered = append(filtered, flow)
					}
				}
				if len(filtered) == 0 {
					return fmt.Errorf("no flows found in namespace '%s'", namespaceFilter)
				}
				parsedFlows = filtered
				fmt.Printf("Filtered to %d flows in namespace '%s'\n", len(parsedFlows), namespaceFilter)
			}

			fmt.Printf("Found %d parsed flows\n", len(parsedFlows))

			// Synthesize policies
			fmt.Println("Synthesizing policies...")
			policies, err := synth.SynthesizePolicies(parsedFlows)
			if err != nil {
				return fmt.Errorf("failed to synthesize policies: %w", err)
			}

			if len(policies) == 0 {
				return fmt.Errorf("no policies generated (flows may be missing required metadata)")
			}

			fmt.Printf("Generated %d policy(ies)\n", len(policies))

			// Write policies to file
			if err := synth.WritePoliciesToFile(policies, outputFile); err != nil {
				return fmt.Errorf("failed to write policies: %w", err)
			}

			fmt.Printf("Policies saved to %s\n", outputFile)

			// Print summary
			for _, policy := range policies {
				fmt.Printf("  - %s/%s (namespace: %s)\n",
					policy.Kind,
					policy.Metadata.Name,
					policy.Metadata.Namespace)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input flows JSON file (default: out/flows.json)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output policy YAML file (default: out/policy.yaml)")
	cmd.Flags().StringVarP(&namespaceFilter, "namespace", "n", "", "Filter flows by namespace (optional)")

	return cmd
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
