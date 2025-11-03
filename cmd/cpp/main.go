package main

import (
    "fmt"
    "os"
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
    return &cobra.Command{
        Use:   "learn",
        Short: "Capture or read Hubble flows",
        RunE: func(cmd *cobra.Command, args []string) error {
            fmt.Println("Learn: reading stub flows.json ...")
            _ = os.MkdirAll("out", 0o755)
            return os.WriteFile("out/flows.json", []byte(`{"schema":"cpp.flows.v1","flows":[]}`), 0o644)
        },
    }
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
            _ = os.MkdirAll("out", 0o755)
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
            _ = os.MkdirAll("out", 0o755)
            html := "# PolicyPilot Report\\n\\n```mermaid\\ngraph TD; frontend-->catalog;\\n```\\n"
            return os.WriteFile("out/report.html", []byte(html), 0o644)
        },
    }
}