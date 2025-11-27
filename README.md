# Cilium PolicyPilot 🐝

Turn real traffic into safe **CiliumNetworkPolicies** in minutes.

PolicyPilot learns from Hubble flows, proposes **least-privilege** policies, verifies them safely, and explains results with diagrams.

## Features

- 🔍 **Learn**: Capture and parse Hubble network flows
- 🎯 **Propose**: Generate least-privilege CiliumNetworkPolicies from observed traffic
- ✅ **Verify**: Validate policy syntax and structure
- 📊 **Explain**: Generate HTML reports with network graphs and statistics

## Quickstart

### Prerequisites

- Go 1.23+ installed
- Access to Hubble flows (JSON format) or Hubble CLI

### Installation

```bash
git clone https://github.com/prabhakaran-jm/cilium-policypilot.git
cd cilium-policypilot
go build -o cpp ./cmd/cpp
```

### Basic Usage

```bash
# 1. Learn from Hubble flows
./cpp learn --input examples/sample-flows.json

# 2. Generate policies
./cpp propose

# 3. Verify policies
./cpp verify

# 4. Generate HTML report
./cpp explain
```

## Commands

### `learn`

Capture or read Hubble flows from JSON files.

```bash
# Read from file
./cpp learn --input flows.json

# Specify output location
./cpp learn --input flows.json --output my-flows.json
```

**Flags:**
- `-i, --input`: Input flows JSON file (default: `out/flows.json`)
- `-o, --output`: Output flows JSON file (default: `out/flows.json`)
- `-d, --duration`: Duration to capture flows (future use)
- `--hubble-endpoint`: Hubble API endpoint (future use)

### `propose`

Generate CiliumNetworkPolicies from parsed flows.

```bash
# Generate policies from default flows file
./cpp propose

# Filter by namespace
./cpp propose --namespace hipstershop

# Custom input/output
./cpp propose --input my-flows.json --output my-policies.yaml
```

**Flags:**
- `-i, --input`: Input flows JSON file (default: `out/flows.json`)
- `-o, --output`: Output policy YAML file (default: `out/policy.yaml`)
- `-n, --namespace`: Filter flows by namespace (optional)

### `verify`

Validate policy YAML syntax and structure.

```bash
# Verify default policy file
./cpp verify

# Verify custom policy file
./cpp verify --input my-policies.yaml
```

**Flags:**
- `-i, --input`: Input policy YAML file (default: `out/policy.yaml`)

**Validates:**
- YAML syntax
- Required fields (apiVersion, kind, metadata, spec)
- CiliumNetworkPolicy structure
- Endpoint selectors
- Ingress/egress rules
- Port and protocol specifications

### `explain`

Generate HTML report with flow statistics, policies, and network visualization.

```bash
# Generate report from default files
./cpp explain

# Custom files
./cpp explain --flows my-flows.json --policies my-policies.yaml --output report.html
```

**Flags:**
- `-f, --flows`: Input flows JSON file (default: `out/flows.json`)
- `-p, --policies`: Input policies YAML file (default: `out/policy.yaml`)
- `-o, --output`: Output HTML report file (default: `out/report.html`)

**Report includes:**
- Statistics dashboard (flows, policies, namespaces, protocols)
- Interactive Mermaid network graph
- Policy list with endpoint selectors
- Namespace and protocol badges

## Examples

### Example 1: Basic Workflow

```bash
# Start with sample flows
./cpp learn --input examples/sample-flows.json

# Generate policies
./cpp propose

# Verify policies
./cpp verify

# Generate report
./cpp explain

# Open report in browser
open out/report.html  # macOS
xdg-open out/report.html  # Linux
start out/report.html  # Windows
```

### Example 2: HipsterShop Microservices

```bash
# Use HipsterShop example
./cpp learn --input examples/hipstershop-flows.json

# Generate policies for hipstershop namespace
./cpp propose --namespace hipstershop --output hipstershop-policies.yaml

# Verify and generate report
./cpp verify --input hipstershop-policies.yaml
./cpp explain --flows out/flows.json --policies hipstershop-policies.yaml
```

### Example 3: Custom Workflow

```bash
# Capture flows from Hubble CLI (if available)
hubble observe -o json --since 5m > my-flows.json

# Process flows
./cpp learn --input my-flows.json --output processed-flows.json
./cpp propose --input processed-flows.json --output my-policies.yaml --namespace production
./cpp verify --input my-policies.yaml
./cpp explain --flows processed-flows.json --policies my-policies.yaml --output my-report.html
```

## Architecture

```
cilium-policypilot/
├── cmd/cpp/              # CLI entry point
├── internal/
│   ├── hubble/          # Hubble flow parsing and reading
│   ├── synth/           # Policy synthesis from flows
│   ├── verify/          # Policy validation
│   ├── explain/         # HTML report generation
│   ├── graph/           # Network graph generation
│   └── validate/        # Input validation utilities
├── examples/            # Example flow files
│   ├── sample-flows.json
│   └── hipstershop-flows.json
└── out/                 # Generated outputs (gitignored)
```

### Data Flow

```
Hubble Flows (JSON)
    ↓
[learn] Parse & Validate
    ↓
Parsed Flows
    ↓
[propose] Synthesize Policies
    ↓
CiliumNetworkPolicy YAML
    ↓
[verify] Validate Structure
    ↓
[explain] Generate Report
    ↓
HTML Report + Network Graph
```

## How It Works

1. **Learn**: Reads Hubble flow data (JSON format) and extracts key metadata:
   - Source/destination pod labels and namespaces
   - Ports and protocols
   - Flow direction and verdict

2. **Propose**: Analyzes flows and generates least-privilege policies:
   - Groups flows by destination endpoint
   - Creates ingress rules from source endpoints
   - Aggregates ports and protocols per source
   - Generates valid CiliumNetworkPolicy YAML

3. **Verify**: Validates generated policies:
   - Checks YAML syntax
   - Validates required fields
   - Ensures proper CiliumNetworkPolicy structure
   - Validates endpoint selectors and rules

4. **Explain**: Creates visual reports:
   - Generates network graph (Mermaid format)
   - Collects statistics (flows, policies, namespaces, protocols)
   - Creates interactive HTML report

## Output Files

- `out/flows.json`: Parsed and validated flows
- `out/policy.yaml`: Generated CiliumNetworkPolicies (multi-document YAML)
- `out/report.html`: HTML report with statistics and network graph

## Why

Writing CiliumNetworkPolicies by hand is error-prone:
- **Too tight**: Breaks workloads, causes outages
- **Too loose**: Opens security holes, violates compliance

PolicyPilot helps engineers find the **sweet spot** by:
- Learning from actual traffic patterns
- Generating least-privilege policies automatically
- Validating policies before deployment
- Visualizing network topology and policies

## Requirements

- **Go**: 1.23 or higher
- **Hubble**: Access to Hubble flows (JSON format) or Hubble CLI
- **Cilium**: Understanding of CiliumNetworkPolicy structure

## Contributing

This project was built for the eBPF Summit Hackathon. Contributions welcome!

## License

Apache License 2.0 - see [LICENSE](LICENSE) file for details.

## Related Projects

- [Cilium](https://github.com/cilium/cilium) - eBPF-based networking, security, and observability
- [Hubble](https://github.com/cilium/hubble) - Network, service & security observability for Kubernetes
- [Tetragon](https://github.com/cilium/tetragon) - eBPF-based Security Observability and Runtime Enforcement

## Acknowledgments

Built for the eBPF Summit Hackathon 2024. Powered by eBPF and Cilium technologies.
