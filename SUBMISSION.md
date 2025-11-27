# Cilium PolicyPilot - Hackathon Submission

## Category: **Cilium Technologies** 🎯

**Category Description:** Use, integrate, enhance or educate about any of the projects under github.com/cilium, including Cilium, Tetragon, Hubble or pwru.

## Why This Project Fits the Category

### ✅ Uses Hubble
- **Primary Data Source**: PolicyPilot consumes Hubble flow data (JSON format)
- **Hubble Integration**: Parses Hubble flow structures and extracts network metadata
- **Future Enhancement**: Planned direct Hubble API integration

### ✅ Generates CiliumNetworkPolicies
- **Output Format**: Produces valid CiliumNetworkPolicy YAML files
- **Cilium-Specific**: Uses Cilium's policy format (not generic Kubernetes NetworkPolicy)
- **Policy Structure**: Generates policies with endpoint selectors, ingress rules, and port/protocol specifications

### ✅ Integrates with Cilium Ecosystem
- **Workflow Integration**: Seamlessly works with Cilium + Hubble setup
- **Policy Deployment**: Generated policies can be directly applied to Cilium-managed clusters
- **Observability**: Leverages Cilium's eBPF-powered observability through Hubble

### ✅ Enhances Cilium Technologies
- **Automation**: Automates the manual process of writing CiliumNetworkPolicies
- **Least-Privilege**: Generates security-focused policies based on actual traffic
- **Validation**: Provides policy verification before deployment
- **Visualization**: Creates network topology graphs from Hubble flows

### ✅ Educates About Cilium
- **Documentation**: Comprehensive guide explaining CiliumNetworkPolicies
- **Examples**: Real-world examples using Cilium technologies
- **Best Practices**: Teaches proper policy creation and deployment
- **Architecture Diagrams**: Visual explanations of Cilium integration

## Project Overview

**Cilium PolicyPilot** is a CLI tool that transforms real network traffic observed by Hubble into secure CiliumNetworkPolicies. It bridges the gap between Cilium's powerful observability (Hubble) and its security capabilities (CiliumNetworkPolicies).

### Key Features

1. **Learn from Hubble**: Captures and parses Hubble network flows
2. **Propose Policies**: Generates least-privilege CiliumNetworkPolicies automatically
3. **Verify**: Validates policy syntax and structure before deployment
4. **Explain**: Creates visual HTML reports with network graphs

### Technology Stack

- **Go 1.23+**: Core language
- **Hubble**: Flow data source (github.com/cilium/hubble)
- **CiliumNetworkPolicies**: Output format (github.com/cilium/cilium)
- **Cobra**: CLI framework
- **Mermaid.js**: Network visualization

## How It Works

```
Hubble Flows (JSON)
    ↓
[learn] Parse & Extract Metadata
    ↓
Parsed Flows
    ↓
[propose] Synthesize CiliumNetworkPolicies
    ↓
CiliumNetworkPolicy YAML
    ↓
[verify] Validate Structure
    ↓
[explain] Generate Visual Report
    ↓
Ready for Cilium Deployment
```

## Use Cases

1. **New Cluster Setup**: Generate initial policies from baseline traffic
2. **Security Hardening**: Create least-privilege policies for existing workloads
3. **Policy Migration**: Convert from permissive to restrictive policies
4. **Documentation**: Visualize network topology and dependencies
5. **Compliance**: Generate policies that meet security requirements

## Project Form

This project is a **code-based tool** that:
- ✅ Provides a working CLI application
- ✅ Includes comprehensive documentation
- ✅ Offers educational resources
- ✅ Demonstrates creative integration of Cilium technologies

## Open Source

- **License**: Apache License 2.0
- **Repository**: https://github.com/prabhakaran-jm/cilium-policypilot
- **Language**: Go
- **Dependencies**: All open source (Cobra, YAML v3, etc.)

## Impact

### For Cilium Users
- **Time Savings**: Reduces policy creation from hours to minutes
- **Security**: Enforces least-privilege by default
- **Accuracy**: Based on real traffic, not assumptions
- **Visibility**: Understands network topology through visualization

### For Cilium Ecosystem
- **Adoption**: Makes CiliumNetworkPolicies more accessible
- **Best Practices**: Promotes security-first policy creation
- **Integration**: Demonstrates powerful Hubble + Cilium workflows
- **Innovation**: Shows creative use of Cilium observability data

## Technical Highlights

- **Modular Architecture**: Clean separation of concerns
- **Comprehensive Testing**: Unit tests for core functionality
- **Error Handling**: Robust validation and error messages
- **Documentation**: Extensive README with diagrams and examples
- **Performance**: Efficient processing of large flow sets

## Future Enhancements

- Direct Hubble API integration
- Egress policy generation
- L7 (HTTP) policy support
- Real-time flow capture
- Policy diff and updates

## Conclusion

Cilium PolicyPilot is a perfect fit for the **Cilium Technologies** category because it:
- Uses Hubble as its primary data source
- Generates CiliumNetworkPolicies as output
- Integrates seamlessly with Cilium ecosystem
- Enhances Cilium's capabilities through automation
- Educates users about CiliumNetworkPolicies

The project demonstrates a creative and practical use of Cilium technologies, making network policy management more accessible and secure for Kubernetes operators.

---

**Built for eBPF Summit Hackathon 2025**  
**Powered by eBPF and Cilium technologies**

