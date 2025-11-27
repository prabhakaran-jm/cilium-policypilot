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

// GenerateGraph creates a network graph from parsed flows.
// Extracts unique nodes (pods) and edges (connections) from flows,
// creating a representation suitable for visualization.
// Aggregates multiple flows between the same nodes into a single edge.
func GenerateGraph(flows []*hubble.ParsedFlow) *Graph {
	graph := &Graph{
		Nodes: make([]Node, 0),
		Edges: make([]Edge, 0),
	}

	// Track unique nodes
	nodeMap := make(map[string]Node)

	// Track edges by source->destination, aggregating ports/protocols
	edgeMap := make(map[string]map[string][]string) // source -> dest -> []protocol:port

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

		// Aggregate edge information
		if edgeMap[sourceID] == nil {
			edgeMap[sourceID] = make(map[string][]string)
		}
		portProto := fmt.Sprintf("%s:%d", flow.Protocol, flow.DestPort)
		// Check if this port/protocol combination already exists
		exists := false
		for _, existing := range edgeMap[sourceID][destID] {
			if existing == portProto {
				exists = true
				break
			}
		}
		if !exists {
			edgeMap[sourceID][destID] = append(edgeMap[sourceID][destID], portProto)
		}
	}

	// Convert node map to slice
	for _, node := range nodeMap {
		graph.Nodes = append(graph.Nodes, node)
	}

	// Convert aggregated edges to Edge slice
	for sourceID, dests := range edgeMap {
		for destID, portProtos := range dests {
			// Aggregate multiple ports/protocols into a single label
			edgeLabel := strings.Join(portProtos, ", ")
			if len(portProtos) > 3 {
				edgeLabel = fmt.Sprintf("%s, ... (%d total)", strings.Join(portProtos[:3], ", "), len(portProtos))
			}

			// Use first port/protocol for the edge struct (for compatibility)
			parts := strings.Split(portProtos[0], ":")
			protocol := parts[0]
			var port uint16
			if len(parts) > 1 {
				fmt.Sscanf(parts[1], "%d", &port)
			}

			edge := Edge{
				From:     sourceID,
				To:       destID,
				Port:     port,
				Protocol: protocol,
				Label:    edgeLabel,
			}
			graph.Edges = append(graph.Edges, edge)
		}
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

// ToMermaid generates a Mermaid diagram string from the graph.
// Returns a Mermaid flowchart syntax string that can be rendered
// in HTML using the Mermaid.js library.
// Limits diagram size to prevent Mermaid "Maximum text size" errors.
func (g *Graph) ToMermaid() string {
	// Mermaid has limits on diagram complexity
	// Limit to reasonable sizes to prevent rendering errors
	maxNodes := 50
	maxEdges := 100

	// If graph is too large, create a simplified version
	if len(g.Nodes) > maxNodes || len(g.Edges) > maxEdges {
		return g.ToMermaidSimplified(maxNodes, maxEdges)
	}

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
		// Escape special characters in edge labels
		edgeLabel = strings.ReplaceAll(edgeLabel, "|", "\\|")
		sb.WriteString(fmt.Sprintf("    %s -->|%s| %s\n", edge.From, edgeLabel, edge.To))
	}

	return sb.String()
}

// ToMermaidSimplified generates a simplified Mermaid diagram for large graphs
func (g *Graph) ToMermaidSimplified(maxNodes, maxEdges int) string {
	var sb strings.Builder

	sb.WriteString("graph TD\n")
	sb.WriteString(fmt.Sprintf("    note1[\"⚠️ Graph Simplified<br/>Too many nodes/edges to display<br/>"))
	sb.WriteString(fmt.Sprintf("Total: %d nodes, %d edges<br/>", len(g.Nodes), len(g.Edges)))
	sb.WriteString(fmt.Sprintf("Showing: %d nodes, %d edges\"]\n", maxNodes, maxEdges))

	// Add limited nodes
	nodeCount := 0
	for _, node := range g.Nodes {
		if nodeCount >= maxNodes {
			break
		}
		nodeLabel := fmt.Sprintf("%s[%s]", node.ID, node.Label)
		if node.Namespace != "" {
			nodeLabel = fmt.Sprintf("%s[%s<br/>ns: %s]", node.ID, node.Label, node.Namespace)
		}
		sb.WriteString(fmt.Sprintf("    %s\n", nodeLabel))
		nodeCount++
	}

	// Add limited edges (only between nodes we're showing)
	nodeSet := make(map[string]bool)
	for i := 0; i < nodeCount && i < len(g.Nodes); i++ {
		nodeSet[g.Nodes[i].ID] = true
	}

	edgeCount := 0
	for _, edge := range g.Edges {
		if edgeCount >= maxEdges {
			break
		}
		if nodeSet[edge.From] && nodeSet[edge.To] {
			edgeLabel := edge.Label
			if edgeLabel == "" {
				edgeLabel = fmt.Sprintf("%s:%d", edge.Protocol, edge.Port)
			}
			edgeLabel = strings.ReplaceAll(edgeLabel, "|", "\\|")
			sb.WriteString(fmt.Sprintf("    %s -->|%s| %s\n", edge.From, edgeLabel, edge.To))
			edgeCount++
		}
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
