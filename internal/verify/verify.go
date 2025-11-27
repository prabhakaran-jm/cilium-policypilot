package verify

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// VerificationResult contains the result of policy verification
type VerificationResult struct {
	Valid    bool
	Errors   []string
	Warnings []string
	Policies []PolicyInfo
}

// PolicyInfo contains information about a verified policy
type PolicyInfo struct {
	Name      string
	Namespace string
	Kind      string
	Valid     bool
	Errors    []string
}

// VerifyPolicies validates policy YAML files for correct syntax and structure.
// Supports multi-document YAML files and validates each policy document.
// Returns a VerificationResult with validation status and detailed error messages.
func VerifyPolicies(filePath string) (*VerificationResult, error) {
	result := &VerificationResult{
		Valid:    true,
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
		Policies: make([]PolicyInfo, 0),
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read policy file: %w", err)
	}

	// Split multi-document YAML
	documents := splitYAMLDocuments(string(data))

	// Verify each document
	for i, doc := range documents {
		if strings.TrimSpace(doc) == "" {
			continue
		}

		policyInfo, err := verifyPolicyDocument(doc, i+1)
		if err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("Document %d: %v", i+1, err))
			result.Policies = append(result.Policies, PolicyInfo{
				Valid:  false,
				Errors: []string{err.Error()},
			})
			continue
		}

		if !policyInfo.Valid {
			result.Valid = false
		}

		result.Policies = append(result.Policies, *policyInfo)
	}

	if len(result.Policies) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "no valid policies found in file")
	}

	return result, nil
}

// verifyPolicyDocument validates a single policy document
func verifyPolicyDocument(yamlDoc string, docNum int) (*PolicyInfo, error) {
	var policy map[string]interface{}

	if err := yaml.Unmarshal([]byte(yamlDoc), &policy); err != nil {
		return nil, fmt.Errorf("invalid YAML syntax: %w", err)
	}

	info := &PolicyInfo{
		Valid:  true,
		Errors: make([]string, 0),
	}

	// Check required top-level fields
	if apiVersion, ok := policy["apiVersion"].(string); ok {
		if apiVersion != "cilium.io/v2" {
			info.Valid = false
			info.Errors = append(info.Errors, fmt.Sprintf("invalid apiVersion: expected 'cilium.io/v2', got '%s'", apiVersion))
		}
	} else {
		info.Valid = false
		info.Errors = append(info.Errors, "missing required field: apiVersion")
	}

	if kind, ok := policy["kind"].(string); ok {
		info.Kind = kind
		if kind != "CiliumNetworkPolicy" {
			info.Valid = false
			info.Errors = append(info.Errors, fmt.Sprintf("invalid kind: expected 'CiliumNetworkPolicy', got '%s'", kind))
		}
	} else {
		info.Valid = false
		info.Errors = append(info.Errors, "missing required field: kind")
	}

	// Check metadata
	if metadata, ok := policy["metadata"].(map[string]interface{}); ok {
		if name, ok := metadata["name"].(string); ok {
			info.Name = name
			if name == "" {
				info.Valid = false
				info.Errors = append(info.Errors, "metadata.name cannot be empty")
			}
		} else {
			info.Valid = false
			info.Errors = append(info.Errors, "missing required field: metadata.name")
		}

		if namespace, ok := metadata["namespace"].(string); ok {
			info.Namespace = namespace
		}
	} else {
		info.Valid = false
		info.Errors = append(info.Errors, "missing required field: metadata")
	}

	// Check spec
	if spec, ok := policy["spec"].(map[string]interface{}); ok {
		// Check endpointSelector
		if endpointSelector, ok := spec["endpointSelector"].(map[string]interface{}); ok {
			if matchLabels, ok := endpointSelector["matchLabels"].(map[string]interface{}); ok {
				if len(matchLabels) == 0 {
					info.Valid = false
					info.Errors = append(info.Errors, "endpointSelector.matchLabels cannot be empty")
				}
			} else {
				info.Valid = false
				info.Errors = append(info.Errors, "missing required field: spec.endpointSelector.matchLabels")
			}
		} else {
			info.Valid = false
			info.Errors = append(info.Errors, "missing required field: spec.endpointSelector")
		}

		// Validate ingress rules if present
		if ingress, ok := spec["ingress"].([]interface{}); ok {
			for i, rule := range ingress {
				if err := validateIngressRule(rule, i); err != nil {
					info.Valid = false
					info.Errors = append(info.Errors, fmt.Sprintf("ingress[%d]: %v", i, err))
				}
			}
		}

		// Validate egress rules if present
		if egress, ok := spec["egress"].([]interface{}); ok {
			for i, rule := range egress {
				if err := validateEgressRule(rule, i); err != nil {
					info.Valid = false
					info.Errors = append(info.Errors, fmt.Sprintf("egress[%d]: %v", i, err))
				}
			}
		}
	} else {
		info.Valid = false
		info.Errors = append(info.Errors, "missing required field: spec")
	}

	return info, nil
}

// validateIngressRule validates an ingress rule
func validateIngressRule(rule interface{}, index int) error {
	ruleMap, ok := rule.(map[string]interface{})
	if !ok {
		return fmt.Errorf("ingress rule must be a map")
	}

	// Check fromEndpoints if present
	if fromEndpoints, ok := ruleMap["fromEndpoints"].([]interface{}); ok {
		for i, ep := range fromEndpoints {
			if epMap, ok := ep.(map[string]interface{}); ok {
				if matchLabels, ok := epMap["matchLabels"].(map[string]interface{}); ok {
					if len(matchLabels) == 0 {
						return fmt.Errorf("fromEndpoints[%d].matchLabels cannot be empty", i)
					}
				} else {
					return fmt.Errorf("fromEndpoints[%d] missing matchLabels", i)
				}
			} else {
				return fmt.Errorf("fromEndpoints[%d] must be a map", i)
			}
		}
	}

	// Check toPorts if present
	if toPorts, ok := ruleMap["toPorts"].([]interface{}); ok {
		for i, portRule := range toPorts {
			if err := validatePortRule(portRule, i); err != nil {
				return fmt.Errorf("toPorts[%d]: %w", i, err)
			}
		}
	}

	return nil
}

// validateEgressRule validates an egress rule
func validateEgressRule(rule interface{}, index int) error {
	ruleMap, ok := rule.(map[string]interface{})
	if !ok {
		return fmt.Errorf("egress rule must be a map")
	}

	// Check toEndpoints if present
	if toEndpoints, ok := ruleMap["toEndpoints"].([]interface{}); ok {
		for i, ep := range toEndpoints {
			if epMap, ok := ep.(map[string]interface{}); ok {
				if matchLabels, ok := epMap["matchLabels"].(map[string]interface{}); ok {
					if len(matchLabels) == 0 {
						return fmt.Errorf("toEndpoints[%d].matchLabels cannot be empty", i)
					}
				} else {
					return fmt.Errorf("toEndpoints[%d] missing matchLabels", i)
				}
			} else {
				return fmt.Errorf("toEndpoints[%d] must be a map", i)
			}
		}
	}

	// Check toPorts if present
	if toPorts, ok := ruleMap["toPorts"].([]interface{}); ok {
		for i, portRule := range toPorts {
			if err := validatePortRule(portRule, i); err != nil {
				return fmt.Errorf("toPorts[%d]: %w", i, err)
			}
		}
	}

	return nil
}

// validatePortRule validates a port rule
func validatePortRule(portRule interface{}, index int) error {
	portRuleMap, ok := portRule.(map[string]interface{})
	if !ok {
		return fmt.Errorf("port rule must be a map")
	}

	ports, ok := portRuleMap["ports"].([]interface{})
	if !ok {
		return fmt.Errorf("missing required field: ports")
	}

	if len(ports) == 0 {
		return fmt.Errorf("ports array cannot be empty")
	}

	for i, port := range ports {
		portMap, ok := port.(map[string]interface{})
		if !ok {
			return fmt.Errorf("ports[%d] must be a map", i)
		}

		// Check port field
		if portVal, ok := portMap["port"].(string); ok {
			if portVal == "" {
				return fmt.Errorf("ports[%d].port cannot be empty", i)
			}
		} else {
			return fmt.Errorf("ports[%d] missing required field: port", i)
		}

		// Check protocol field
		if protocol, ok := portMap["protocol"].(string); ok {
			validProtocols := map[string]bool{
				"TCP":  true,
				"UDP":  true,
				"ICMP": true,
				"SCTP": true,
			}
			if !validProtocols[strings.ToUpper(protocol)] {
				return fmt.Errorf("ports[%d].protocol invalid: must be TCP, UDP, ICMP, or SCTP", i)
			}
		} else {
			return fmt.Errorf("ports[%d] missing required field: protocol", i)
		}
	}

	return nil
}

// splitYAMLDocuments splits multi-document YAML into individual documents
func splitYAMLDocuments(yamlContent string) []string {
	documents := make([]string, 0)
	currentDoc := strings.Builder{}

	lines := strings.Split(yamlContent, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "---" {
			if currentDoc.Len() > 0 {
				documents = append(documents, currentDoc.String())
				currentDoc.Reset()
			}
			continue
		}
		currentDoc.WriteString(line)
		currentDoc.WriteString("\n")
	}

	if currentDoc.Len() > 0 {
		documents = append(documents, currentDoc.String())
	}

	return documents
}
