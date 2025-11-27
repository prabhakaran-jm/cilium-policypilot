package synth

import (
	"fmt"
	"sort"
	"strings"

	"github.com/prabhakaran-jm/cilium-policypilot/internal/hubble"
)

// Policy represents a CiliumNetworkPolicy
type Policy struct {
	APIVersion string         `yaml:"apiVersion"`
	Kind       string         `yaml:"kind"`
	Metadata   PolicyMetadata `yaml:"metadata"`
	Spec       PolicySpec     `yaml:"spec"`
}

// PolicyMetadata contains policy metadata
type PolicyMetadata struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace,omitempty"`
}

// PolicySpec contains the policy specification
type PolicySpec struct {
	EndpointSelector EndpointSelector `yaml:"endpointSelector"`
	Ingress          []IngressRule    `yaml:"ingress,omitempty"`
	Egress           []EgressRule     `yaml:"egress,omitempty"`
}

// EndpointSelector selects endpoints for the policy
type EndpointSelector struct {
	MatchLabels map[string]string `yaml:"matchLabels"`
}

// IngressRule defines an ingress rule
type IngressRule struct {
	FromEndpoints []EndpointSelector `yaml:"fromEndpoints,omitempty"`
	ToPorts       []PortRule         `yaml:"toPorts,omitempty"`
}

// EgressRule defines an egress rule
type EgressRule struct {
	ToEndpoints []EndpointSelector `yaml:"toEndpoints,omitempty"`
	ToPorts     []PortRule         `yaml:"toPorts,omitempty"`
}

// PortRule defines port and protocol rules
type PortRule struct {
	Ports []PortProtocol `yaml:"ports"`
}

// PortProtocol defines a port and protocol
type PortProtocol struct {
	Port     string `yaml:"port"`
	Protocol string `yaml:"protocol"`
}

// EndpointKey uniquely identifies an endpoint for grouping flows
type EndpointKey struct {
	Namespace string
	Labels    map[string]string
}

// EndpointFlows groups flows by destination endpoint
type EndpointFlows struct {
	Key   EndpointKey
	Flows []*hubble.ParsedFlow
}

// SynthesizePolicies generates CiliumNetworkPolicies from parsed flows
func SynthesizePolicies(flows []*hubble.ParsedFlow) ([]*Policy, error) {
	if len(flows) == 0 {
		return nil, fmt.Errorf("no flows provided")
	}

	// Group flows by destination endpoint
	endpointGroups := groupFlowsByEndpoint(flows)

	// Generate policies for each endpoint group
	policies := make([]*Policy, 0, len(endpointGroups))
	for _, group := range endpointGroups {
		policy, err := generatePolicyForEndpoint(group)
		if err != nil {
			return nil, fmt.Errorf("failed to generate policy for endpoint: %w", err)
		}
		if policy != nil {
			policies = append(policies, policy)
		}
	}

	return policies, nil
}

// groupFlowsByEndpoint groups flows by their destination endpoint
func groupFlowsByEndpoint(flows []*hubble.ParsedFlow) []*EndpointFlows {
	groups := make(map[string]*EndpointFlows)

	for _, flow := range flows {
		// Skip flows without destination information
		if flow.DestNamespace == "" || len(flow.DestLabels) == 0 {
			continue
		}

		// Create key for destination endpoint
		key := EndpointKey{
			Namespace: flow.DestNamespace,
			Labels:    flow.DestLabels,
		}

		// Create string key for map lookup
		keyStr := endpointKeyToString(key)

		// Get or create group for this endpoint
		group, exists := groups[keyStr]
		if !exists {
			group = &EndpointFlows{
				Key:   key,
				Flows: make([]*hubble.ParsedFlow, 0),
			}
			groups[keyStr] = group
		}

		group.Flows = append(group.Flows, flow)
	}

	// Convert map to slice
	result := make([]*EndpointFlows, 0, len(groups))
	for _, group := range groups {
		result = append(result, group)
	}

	// Sort by namespace and labels for consistent output
	sort.Slice(result, func(i, j int) bool {
		if result[i].Key.Namespace != result[j].Key.Namespace {
			return result[i].Key.Namespace < result[j].Key.Namespace
		}
		// Simple comparison of label keys (could be improved)
		return fmt.Sprintf("%v", result[i].Key.Labels) < fmt.Sprintf("%v", result[j].Key.Labels)
	})

	return result
}

// endpointKeyToString converts an EndpointKey to a string for map key usage
func endpointKeyToString(key EndpointKey) string {
	// Create a deterministic string representation
	labelPairs := make([]string, 0, len(key.Labels))
	for k, v := range key.Labels {
		labelPairs = append(labelPairs, fmt.Sprintf("%s=%s", k, v))
	}
	// Sort for consistency
	sort.Strings(labelPairs)
	return fmt.Sprintf("%s:%s", key.Namespace, strings.Join(labelPairs, ","))
}

// generatePolicyForEndpoint generates a policy for a specific endpoint group
func generatePolicyForEndpoint(group *EndpointFlows) (*Policy, error) {
	if len(group.Flows) == 0 {
		return nil, nil
	}

	// Extract app label for policy name, fallback to first label key
	policyName := generatePolicyName(group.Key.Labels)

	// Generate ingress rules from flows
	ingressRules := generateIngressRules(group.Flows)

	// Only create policy if we have ingress rules
	if len(ingressRules) == 0 {
		return nil, nil
	}

	policy := &Policy{
		APIVersion: "cilium.io/v2",
		Kind:       "CiliumNetworkPolicy",
		Metadata: PolicyMetadata{
			Name:      policyName,
			Namespace: group.Key.Namespace,
		},
		Spec: PolicySpec{
			EndpointSelector: EndpointSelector{
				MatchLabels: group.Key.Labels,
			},
			Ingress: ingressRules,
		},
	}

	return policy, nil
}

// generatePolicyName creates a policy name from endpoint labels
func generatePolicyName(labels map[string]string) string {
	// Try to find common label keys
	preferredKeys := []string{"app", "k8s:app", "name", "component"}

	for _, key := range preferredKeys {
		if value, exists := labels[key]; exists {
			return fmt.Sprintf("%s-policy", value)
		}
	}

	// Fallback to first label value
	for _, value := range labels {
		return fmt.Sprintf("%s-policy", value)
	}

	return "default-policy"
}

// generateIngressRules creates ingress rules from flows
func generateIngressRules(flows []*hubble.ParsedFlow) []IngressRule {
	// Group flows by source endpoint and port/protocol
	ruleMap := make(map[string]*IngressRule)

	for _, flow := range flows {
		// Skip flows without source information
		if len(flow.SourceLabels) == 0 {
			continue
		}

		// Skip flows without port information
		if flow.DestPort == 0 {
			continue
		}

		// Create a key for grouping: source labels + port + protocol
		// We'll group by source endpoint first, then combine ports
		sourceKey := fmt.Sprintf("%v", flow.SourceLabels)

		rule, exists := ruleMap[sourceKey]
		if !exists {
			rule = &IngressRule{
				FromEndpoints: []EndpointSelector{
					{MatchLabels: flow.SourceLabels},
				},
				ToPorts: []PortRule{},
			}
			ruleMap[sourceKey] = rule
		}

		// Add port if not already present
		portStr := fmt.Sprintf("%d", flow.DestPort)
		protocol := flow.Protocol
		if protocol == "" {
			protocol = "TCP"
		}

		portExists := false
		for _, portRule := range rule.ToPorts {
			for _, pp := range portRule.Ports {
				if pp.Port == portStr && pp.Protocol == protocol {
					portExists = true
					break
				}
			}
			if portExists {
				break
			}
		}

		if !portExists {
			// Find or create PortRule for this protocol
			portRuleIndex := -1
			for i, pr := range rule.ToPorts {
				if len(pr.Ports) > 0 && pr.Ports[0].Protocol == protocol {
					portRuleIndex = i
					break
				}
			}

			if portRuleIndex >= 0 {
				// Add port to existing PortRule
				rule.ToPorts[portRuleIndex].Ports = append(rule.ToPorts[portRuleIndex].Ports, PortProtocol{
					Port:     portStr,
					Protocol: protocol,
				})
			} else {
				// Create new PortRule
				rule.ToPorts = append(rule.ToPorts, PortRule{
					Ports: []PortProtocol{
						{
							Port:     portStr,
							Protocol: protocol,
						},
					},
				})
			}
		}
	}

	// Convert map to slice
	rules := make([]IngressRule, 0, len(ruleMap))
	for _, rule := range ruleMap {
		// Sort ports within each rule
		for i := range rule.ToPorts {
			sort.Slice(rule.ToPorts[i].Ports, func(a, b int) bool {
				return rule.ToPorts[i].Ports[a].Port < rule.ToPorts[i].Ports[b].Port
			})
		}
		rules = append(rules, *rule)
	}

	// Sort rules by source labels for consistent output
	sort.Slice(rules, func(i, j int) bool {
		return fmt.Sprintf("%v", rules[i].FromEndpoints[0].MatchLabels) <
			fmt.Sprintf("%v", rules[j].FromEndpoints[0].MatchLabels)
	})

	return rules
}
