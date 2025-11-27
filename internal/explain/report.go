package explain

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/prabhakaran-jm/cilium-policypilot/internal/graph"
	"github.com/prabhakaran-jm/cilium-policypilot/internal/hubble"
	"github.com/prabhakaran-jm/cilium-policypilot/internal/synth"
)

// ReportData contains data for generating the report
type ReportData struct {
	GeneratedAt     time.Time
	FlowCount       int
	ParsedFlowCount int
	PolicyCount     int
	Policies        []*synth.Policy
	Graph           *graph.Graph
	Namespaces      []string
	Protocols       map[string]int
}

// GenerateReport generates an HTML report from flows and policies
func GenerateReport(flows []*hubble.ParsedFlow, policies []*synth.Policy) (*ReportData, error) {
	// Generate network graph
	networkGraph := graph.GenerateGraph(flows)

	// Collect statistics
	namespaces := collectNamespaces(flows)
	protocols := collectProtocols(flows)

	data := &ReportData{
		GeneratedAt:     time.Now(),
		FlowCount:       len(flows),
		ParsedFlowCount: len(flows),
		PolicyCount:     len(policies),
		Policies:        policies,
		Graph:           networkGraph,
		Namespaces:      namespaces,
		Protocols:       protocols,
	}

	return data, nil
}

// WriteHTMLReport writes an HTML report to a file
func WriteHTMLReport(data *ReportData, filePath string) error {
	html := generateHTML(data)

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(filePath, []byte(html), 0644); err != nil {
		return fmt.Errorf("failed to write HTML report: %w", err)
	}

	return nil
}

// generateHTML creates the HTML content
func generateHTML(data *ReportData) string {
	var sb strings.Builder

	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>PolicyPilot Report</title>
    <script src="https://cdn.jsdelivr.net/npm/mermaid/dist/mermaid.min.js"></script>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            border-radius: 10px;
            margin-bottom: 30px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        .header h1 {
            margin: 0;
            font-size: 2.5em;
        }
        .header p {
            margin: 10px 0 0 0;
            opacity: 0.9;
        }
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .stat-card {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .stat-card h3 {
            margin: 0 0 10px 0;
            color: #666;
            font-size: 0.9em;
            text-transform: uppercase;
        }
        .stat-card .value {
            font-size: 2em;
            font-weight: bold;
            color: #667eea;
        }
        .section {
            background: white;
            padding: 30px;
            border-radius: 8px;
            margin-bottom: 30px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .section h2 {
            margin-top: 0;
            color: #333;
            border-bottom: 2px solid #667eea;
            padding-bottom: 10px;
        }
        .policy-list {
            list-style: none;
            padding: 0;
        }
        .policy-item {
            background: #f8f9fa;
            padding: 15px;
            margin: 10px 0;
            border-radius: 5px;
            border-left: 4px solid #667eea;
        }
        .policy-item strong {
            color: #667eea;
        }
        .mermaid {
            background: white;
            padding: 20px;
            border-radius: 8px;
            margin: 20px 0;
        }
        .protocol-list {
            display: flex;
            flex-wrap: wrap;
            gap: 10px;
        }
        .protocol-badge {
            background: #667eea;
            color: white;
            padding: 5px 15px;
            border-radius: 20px;
            font-size: 0.9em;
        }
        .namespace-list {
            display: flex;
            flex-wrap: wrap;
            gap: 10px;
        }
        .namespace-badge {
            background: #764ba2;
            color: white;
            padding: 5px 15px;
            border-radius: 20px;
            font-size: 0.9em;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>🐝 PolicyPilot Report</h1>
        <p>Generated at ` + data.GeneratedAt.Format("2006-01-02 15:04:05 MST") + `</p>
    </div>

    <div class="stats">
        <div class="stat-card">
            <h3>Total Flows</h3>
            <div class="value">` + fmt.Sprintf("%d", data.FlowCount) + `</div>
        </div>
        <div class="stat-card">
            <h3>Policies Generated</h3>
            <div class="value">` + fmt.Sprintf("%d", data.PolicyCount) + `</div>
        </div>
        <div class="stat-card">
            <h3>Namespaces</h3>
            <div class="value">` + fmt.Sprintf("%d", len(data.Namespaces)) + `</div>
        </div>
        <div class="stat-card">
            <h3>Protocols</h3>
            <div class="value">` + fmt.Sprintf("%d", len(data.Protocols)) + `</div>
        </div>
    </div>

    <div class="section">
        <h2>📊 Network Graph</h2>
        <div class="mermaid">
` + data.Graph.ToMermaid() + `
        </div>
    </div>

    <div class="section">
        <h2>📋 Generated Policies</h2>
        <ul class="policy-list">`)

	for _, policy := range data.Policies {
		sb.WriteString(fmt.Sprintf(`
            <li class="policy-item">
                <strong>%s</strong> (namespace: %s)
                <br>
                <small>Protects endpoints matching: %s</small>
            </li>`,
			policy.Metadata.Name,
			policy.Metadata.Namespace,
			formatLabels(policy.Spec.EndpointSelector.MatchLabels)))
	}

	sb.WriteString(`
        </ul>
    </div>

    <div class="section">
        <h2>🌐 Namespaces</h2>
        <div class="namespace-list">`)

	for _, ns := range data.Namespaces {
		sb.WriteString(fmt.Sprintf(`<span class="namespace-badge">%s</span>`, ns))
	}

	sb.WriteString(`
        </div>
    </div>

    <div class="section">
        <h2>🔌 Protocols</h2>
        <div class="protocol-list">`)

	for protocol, count := range data.Protocols {
		sb.WriteString(fmt.Sprintf(`<span class="protocol-badge">%s: %d</span>`, protocol, count))
	}

	sb.WriteString(`
        </div>
    </div>

    <script>
        mermaid.initialize({ startOnLoad: true, theme: 'default' });
    </script>
</body>
</html>`)

	return sb.String()
}

// collectNamespaces extracts unique namespaces from flows
func collectNamespaces(flows []*hubble.ParsedFlow) []string {
	nsMap := make(map[string]bool)
	for _, flow := range flows {
		if flow.SourceNamespace != "" {
			nsMap[flow.SourceNamespace] = true
		}
		if flow.DestNamespace != "" {
			nsMap[flow.DestNamespace] = true
		}
	}

	namespaces := make([]string, 0, len(nsMap))
	for ns := range nsMap {
		namespaces = append(namespaces, ns)
	}
	sort.Strings(namespaces)
	return namespaces
}

// collectProtocols counts protocols used in flows
func collectProtocols(flows []*hubble.ParsedFlow) map[string]int {
	protocols := make(map[string]int)
	for _, flow := range flows {
		if flow.Protocol != "" {
			protocols[flow.Protocol]++
		}
	}
	return protocols
}

// formatLabels formats labels map as a string
func formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return "none"
	}

	pairs := make([]string, 0, len(labels))
	for k, v := range labels {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(pairs)
	return strings.Join(pairs, ", ")
}
