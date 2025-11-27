package graph

import (
	"fmt"
	"sort"
	"strings"

	"github.com/prabhakaran-jm/cilium-policypilot/internal/hubble"
)

// Node represents a node in the network graph
type Node struct {
	ID        string
	Label     string
	Namespace string
	Type      string // "pod", "service", etc.
}

// Edge represents a connection between nodes
type Edge struct {
	From     string
	To       string
	Port     uint16
	Protocol string
	Label    string
}

// Graph represents a network graph
type Graph struct {
	Nodes []Node
	Edges []Edge
}

// GenerateGraph creates a network graph from parsed flows
func GenerateGraph(flows []*hubble.ParsedFlow) *Graph {
	graph := &Graph{
		Nodes: make([]Node, 0),
		Edges: make([]Edge, 0),
	}

	// Track unique nodes
	nodeMap := make(map[string]Node)

	// Process flows to extract nodes and edges
	for _, flow := range flows {
		// Skip flows without proper source/destination
		if len(flow.SourceLabels) == 0 || len(flow.DestLabels) == 0 {
			continue
		}

		// Create or get source node
		sourceID := getNodeID(flow.SourceLabels, flow.SourceNamespace)
		if _, exists := nodeMap[sourceID]; !exists {
			sourceNode := Node{
				ID:        sourceID,
				Label:     getNodeLabel(flow.SourceLabels),
				Namespace: flow.SourceNamespace,
				Type:      "pod",
			}
			nodeMap[sourceID] = sourceNode
		}

		// Create or get destination node
		destID := getNodeID(flow.DestLabels, flow.DestNamespace)
		if _, exists := nodeMap[destID]; !exists {
			destNode := Node{
				ID:        destID,
				Label:     getNodeLabel(flow.DestLabels),
				Namespace: flow.DestNamespace,
				Type:      "pod",
			}
			nodeMap[destID] = destNode
		}

		// Create edge
		edgeLabel := fmt.Sprintf("%s:%d", flow.Protocol, flow.DestPort)
		edge := Edge{
			From:     sourceID,
			To:       destID,
			Port:     flow.DestPort,
			Protocol: flow.Protocol,
			Label:    edgeLabel,
		}
		graph.Edges = append(graph.Edges, edge)
	}

	// Convert node map to slice
	for _, node := range nodeMap {
		graph.Nodes = append(graph.Nodes, node)
	}

	// Sort nodes and edges for consistent output
	sort.Slice(graph.Nodes, func(i, j int) bool {
		return graph.Nodes[i].ID < graph.Nodes[j].ID
	})
	sort.Slice(graph.Edges, func(i, j int) bool {
		if graph.Edges[i].From != graph.Edges[j].From {
			return graph.Edges[i].From < graph.Edges[j].From
		}
		return graph.Edges[i].To < graph.Edges[j].To
	})

	return graph
}

// ToMermaid generates a Mermaid diagram string from the graph
func (g *Graph) ToMermaid() string {
	var sb strings.Builder

	sb.WriteString("graph TD\n")

	// Add nodes
	for _, node := range g.Nodes {
		nodeLabel := fmt.Sprintf("%s[%s]", node.ID, node.Label)
		if node.Namespace != "" {
			nodeLabel = fmt.Sprintf("%s[%s<br/>ns: %s]", node.ID, node.Label, node.Namespace)
		}
		sb.WriteString(fmt.Sprintf("    %s\n", nodeLabel))
	}

	// Add edges
	for _, edge := range g.Edges {
		edgeLabel := edge.Label
		if edgeLabel == "" {
			edgeLabel = fmt.Sprintf("%s:%d", edge.Protocol, edge.Port)
		}
		sb.WriteString(fmt.Sprintf("    %s -->|%s| %s\n", edge.From, edgeLabel, edge.To))
	}

	return sb.String()
}

// getNodeID creates a unique ID for a node based on labels and namespace
func getNodeID(labels map[string]string, namespace string) string {
	// Try to find app label first
	if app, exists := labels["k8s:app"]; exists {
		return sanitizeID(fmt.Sprintf("%s-%s", namespace, app))
	}
	if app, exists := labels["app"]; exists {
		return sanitizeID(fmt.Sprintf("%s-%s", namespace, app))
	}

	// Fallback to first label value
	for _, value := range labels {
		return sanitizeID(fmt.Sprintf("%s-%s", namespace, value))
	}

	return sanitizeID(namespace)
}

// getNodeLabel extracts a human-readable label from pod labels
func getNodeLabel(labels map[string]string) string {
	// Try common label keys
	preferredKeys := []string{"k8s:app", "app", "name", "component"}
	for _, key := range preferredKeys {
		if value, exists := labels[key]; exists {
			return value
		}
	}

	// Fallback to first label value
	for _, value := range labels {
		return value
	}

	return "unknown"
}

// sanitizeID sanitizes a string to be used as a Mermaid node ID
func sanitizeID(id string) string {
	// Replace invalid characters with hyphens
	id = strings.ToLower(id)
	id = strings.ReplaceAll(id, ":", "-")
	id = strings.ReplaceAll(id, ".", "-")
	id = strings.ReplaceAll(id, "_", "-")
	id = strings.ReplaceAll(id, " ", "-")

	// Remove consecutive hyphens
	for strings.Contains(id, "--") {
		id = strings.ReplaceAll(id, "--", "-")
	}

	// Remove leading/trailing hyphens
	id = strings.Trim(id, "-")

	return id
}

