package hubble

import (
	"encoding/json"
	"fmt"
	"os"
)

// ReadFlowsFromFile reads and parses flows from a JSON file
func ReadFlowsFromFile(filePath string) (*FlowCollection, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read flows file: %w", err)
	}

	var collection FlowCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, fmt.Errorf("failed to parse flows JSON: %w", err)
	}

	return &collection, nil
}

// ParseFlow extracts key metadata from a Flow for policy generation
func ParseFlow(flow *Flow) (*ParsedFlow, error) {
	if flow == nil {
		return nil, fmt.Errorf("flow is nil")
	}

	parsed := &ParsedFlow{
		SourceLabels:    make(map[string]string),
		DestLabels:      make(map[string]string),
		SourceNamespace: "",
		DestNamespace:   "",
		Protocol:        "TCP",     // default
		Direction:       "ingress", // default from destination perspective
		Verdict:         flow.Verdict,
	}

	// Extract source endpoint information
	if flow.Source != nil {
		parsed.SourceLabels = ParseLabels(flow.Source.Labels)
		parsed.SourceNamespace = flow.Source.Namespace
		parsed.SourcePod = flow.Source.PodName
	}

	// Extract destination endpoint information
	if flow.Destination != nil {
		parsed.DestLabels = ParseLabels(flow.Destination.Labels)
		parsed.DestNamespace = flow.Destination.Namespace
		parsed.DestPod = flow.Destination.PodName
	}

	// Extract transport layer information
	if flow.L4 != nil {
		if flow.L4.TCP != nil {
			parsed.Protocol = "TCP"
			parsed.DestPort = flow.L4.TCP.DestinationPort
		} else if flow.L4.UDP != nil {
			parsed.Protocol = "UDP"
			parsed.DestPort = flow.L4.UDP.DestinationPort
		}
	}

	// Determine direction: if we have both source and dest, it's ingress to destination
	// For now, we'll treat flows as ingress to the destination pod
	if parsed.DestPod != "" {
		parsed.Direction = "ingress"
	}

	return parsed, nil
}

// ParseFlows extracts metadata from all flows in a collection
func ParseFlows(collection *FlowCollection) ([]*ParsedFlow, error) {
	if collection == nil {
		return nil, fmt.Errorf("flow collection is nil")
	}

	parsedFlows := make([]*ParsedFlow, 0, len(collection.Flows))
	for _, flow := range collection.Flows {
		parsed, err := ParseFlow(flow)
		if err != nil {
			// Log error but continue processing other flows
			continue
		}
		parsedFlows = append(parsedFlows, parsed)
	}

	return parsedFlows, nil
}

// WriteFlowsToFile writes a FlowCollection to a JSON file
func WriteFlowsToFile(collection *FlowCollection, filePath string) error {
	data, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal flows: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write flows file: %w", err)
	}

	return nil
}
