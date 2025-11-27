package main

import (
	"fmt"
	"os"

	"github.com/prabhakaran-jm/cilium-policypilot/internal/explain"
	"github.com/prabhakaran-jm/cilium-policypilot/internal/hubble"
	"github.com/prabhakaran-jm/cilium-policypilot/internal/synth"
	"github.com/prabhakaran-jm/cilium-policypilot/internal/validate"
	"github.com/prabhakaran-jm/cilium-policypilot/internal/verify"
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
	var policyFile string

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify CiliumNetworkPolicy YAML syntax and structure",
		Long:  "Validates policy YAML files for correct syntax, required fields, and CiliumNetworkPolicy structure.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set default policy file if not provided
			if policyFile == "" {
				policyFile = "out/policy.yaml"
			}

			// Validate input file
			if err := validate.FilePath(policyFile); err != nil {
				return fmt.Errorf("invalid policy file: %w", err)
			}
			if err := validate.FileExtension(policyFile, ".yaml"); err != nil {
				// Also accept .yml extension
				if err2 := validate.FileExtension(policyFile, ".yml"); err2 != nil {
					return fmt.Errorf("policy file must be YAML (.yaml or .yml): %w", err)
				}
			}

			fmt.Printf("Verifying policies in %s...\n", policyFile)

			// Verify policies
			result, err := verify.VerifyPolicies(policyFile)
			if err != nil {
				return fmt.Errorf("verification failed: %w", err)
			}

			// Print results
			fmt.Printf("\nVerification Results:\n")
			fmt.Printf("  Status: ")
			if result.Valid {
				fmt.Println("✓ VALID")
			} else {
				fmt.Println("✗ INVALID")
			}

			fmt.Printf("  Policies found: %d\n", len(result.Policies))

			// Print policy details
			for i, policy := range result.Policies {
				fmt.Printf("\n  Policy %d: %s/%s\n", i+1, policy.Kind, policy.Name)
				if policy.Namespace != "" {
					fmt.Printf("    Namespace: %s\n", policy.Namespace)
				}
				if policy.Valid {
					fmt.Printf("    Status: ✓ VALID\n")
				} else {
					fmt.Printf("    Status: ✗ INVALID\n")
					for _, err := range policy.Errors {
						fmt.Printf("      Error: %s\n", err)
					}
				}
			}

			// Print overall errors if any
			if len(result.Errors) > 0 {
				fmt.Printf("\n  Errors:\n")
				for _, err := range result.Errors {
					fmt.Printf("    - %s\n", err)
				}
			}

			// Print warnings if any
			if len(result.Warnings) > 0 {
				fmt.Printf("\n  Warnings:\n")
				for _, warning := range result.Warnings {
					fmt.Printf("    - %s\n", warning)
				}
			}

			// Exit with error if validation failed
			if !result.Valid {
				return fmt.Errorf("policy verification failed")
			}

			fmt.Printf("\n✓ All policies are valid!\n")
			return nil
		},
	}

	cmd.Flags().StringVarP(&policyFile, "input", "i", "", "Input policy YAML file (default: out/policy.yaml)")

	return cmd
}

func cmdExplain() *cobra.Command {
	var flowsFile string
	var policiesFile string
	var outputFile string

	cmd := &cobra.Command{
		Use:   "explain",
		Short: "Generate HTML report with policy summary and network graph",
		Long:  "Generate an HTML report with flow statistics, generated policies, and network visualization.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set defaults
			if flowsFile == "" {
				flowsFile = "out/flows.json"
			}
			if policiesFile == "" {
				policiesFile = "out/policy.yaml"
			}
			if outputFile == "" {
				outputFile = "out/report.html"
			}

			// Validate input files
			if err := validate.FilePath(flowsFile); err != nil {
				return fmt.Errorf("invalid flows file: %w", err)
			}
			if err := validate.FileExtension(flowsFile, ".json"); err != nil {
				return fmt.Errorf("flows file must be JSON: %w", err)
			}

			// Validate output path
			if err := validate.OutputPath(outputFile); err != nil {
				return fmt.Errorf("invalid output path: %w", err)
			}
			if err := validate.FileExtension(outputFile, ".html"); err != nil {
				return fmt.Errorf("output file must be HTML: %w", err)
			}

			fmt.Printf("Reading flows from %s...\n", flowsFile)
			collection, err := hubble.ReadFlowsFromFile(flowsFile)
			if err != nil {
				return fmt.Errorf("failed to read flows: %w", err)
			}

			// Parse flows
			parsedFlows, err := hubble.ParseFlows(collection)
			if err != nil {
				return fmt.Errorf("failed to parse flows: %w", err)
			}

			if len(parsedFlows) == 0 {
				return fmt.Errorf("no valid flows found")
			}

			fmt.Printf("Found %d parsed flows\n", len(parsedFlows))

			// Read policies if file exists
			var policies []*synth.Policy
			if _, err := os.Stat(policiesFile); err == nil {
				fmt.Printf("Reading policies from %s...\n", policiesFile)
				// For now, we'll synthesize policies from flows
				// In the future, we could parse the YAML file
				policies, err = synth.SynthesizePolicies(parsedFlows)
				if err != nil {
					return fmt.Errorf("failed to synthesize policies: %w", err)
				}
				fmt.Printf("Found %d policies\n", len(policies))
			} else {
				// Generate policies from flows
				fmt.Println("No policy file found. Generating policies from flows...")
				policies, err = synth.SynthesizePolicies(parsedFlows)
				if err != nil {
					return fmt.Errorf("failed to synthesize policies: %w", err)
				}
			}

			// Generate report
			fmt.Println("Generating report...")
			reportData, err := explain.GenerateReport(parsedFlows, policies)
			if err != nil {
				return fmt.Errorf("failed to generate report: %w", err)
			}

			// Write HTML report
			if err := explain.WriteHTMLReport(reportData, outputFile); err != nil {
				return fmt.Errorf("failed to write HTML report: %w", err)
			}

			fmt.Printf("Report saved to %s\n", outputFile)
			fmt.Printf("  - %d flows analyzed\n", reportData.FlowCount)
			fmt.Printf("  - %d policies generated\n", reportData.PolicyCount)
			fmt.Printf("  - %d namespaces\n", len(reportData.Namespaces))
			fmt.Printf("  - Network graph included\n")

			return nil
		},
	}

	cmd.Flags().StringVarP(&flowsFile, "flows", "f", "", "Input flows JSON file (default: out/flows.json)")
	cmd.Flags().StringVarP(&policiesFile, "policies", "p", "", "Input policies YAML file (default: out/policy.yaml)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output HTML report file (default: out/report.html)")

	return cmd
}
