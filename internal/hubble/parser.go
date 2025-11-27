package hubble

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ReadFlowsFromFile reads and parses flows from a JSON file.
// Supports both PolicyPilot format (single JSON object with flows array)
// and Hubble NDJSON format (newline-delimited JSON with flow objects).
func ReadFlowsFromFile(filePath string) (*FlowCollection, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read flows file: %w", err)
	}

	// Try parsing as single JSON object first (PolicyPilot format)
	// Normalize field names first: "IP" -> "ip", "ipVersion" string -> int
	dataStr := string(data)
	dataStr = strings.ReplaceAll(dataStr, `"IP":`, `"ip":`)
	dataStr = strings.ReplaceAll(dataStr, `"ipVersion":"IPv4"`, `"ipVersion":4`)
	dataStr = strings.ReplaceAll(dataStr, `"ipVersion":"IPv6"`, `"ipVersion":6`)

	// Try unmarshaling into FlowCollection
	var collection FlowCollection
	if err := json.Unmarshal([]byte(dataStr), &collection); err == nil && collection.Schema != "" {
		return &collection, nil
	}

	// If that failed, try a more lenient approach: unmarshal into map and convert
	// This handles cases where the JSON has extra fields that don't match the struct
	var rawCollection map[string]interface{}
	if err2 := json.Unmarshal([]byte(dataStr), &rawCollection); err2 == nil {
		if schema, ok := rawCollection["schema"].(string); ok && schema != "" {
			if flowsRaw, ok := rawCollection["flows"].([]interface{}); ok {
				flows := make([]*Flow, 0, len(flowsRaw))
				for _, flowRaw := range flowsRaw {
					flowJSON, _ := json.Marshal(flowRaw)
					var flow Flow
					// Use json.Unmarshal with strict mode disabled - it will ignore unknown fields
					if err3 := json.Unmarshal(flowJSON, &flow); err3 == nil {
						flows = append(flows, &flow)
					}
				}
				if len(flows) > 0 {
					return &FlowCollection{
						Schema: schema,
						Flows:  flows,
					}, nil
				}
			}
		}
	}

	// If that fails, try parsing as NDJSON (Hubble format)
	// Each line is: {"flow":{...},"node_name":"...","time":"..."}
	lines := strings.Split(string(data), "\n")
	flows := make([]*Flow, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse line as JSON
		var lineObj map[string]interface{}
		if err := json.Unmarshal([]byte(line), &lineObj); err != nil {
			continue // Skip invalid lines
		}

		// Extract flow object
		if flowData, ok := lineObj["flow"]; ok {
			flowJSON, err := json.Marshal(flowData)
			if err != nil {
				continue
			}

			// Normalize field names: "IP" -> "ip", handle ipVersion string -> int
			flowJSONStr := string(flowJSON)
			flowJSONStr = strings.ReplaceAll(flowJSONStr, `"IP":`, `"ip":`)

			// Convert ipVersion string to int if needed
			if strings.Contains(flowJSONStr, `"ipVersion":"IPv4"`) {
				flowJSONStr = strings.ReplaceAll(flowJSONStr, `"ipVersion":"IPv4"`, `"ipVersion":4`)
			} else if strings.Contains(flowJSONStr, `"ipVersion":"IPv6"`) {
				flowJSONStr = strings.ReplaceAll(flowJSONStr, `"ipVersion":"IPv6"`, `"ipVersion":6`)
			}

			var flow Flow
			if err := json.Unmarshal([]byte(flowJSONStr), &flow); err == nil {
				flows = append(flows, &flow)
			}
		}
	}

	if len(flows) > 0 {
		return &FlowCollection{
			Schema: "cpp.flows.v1",
			Flows:  flows,
		}, nil
	}

	return nil, fmt.Errorf("failed to parse flows JSON: could not parse as single JSON or NDJSON format")
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
