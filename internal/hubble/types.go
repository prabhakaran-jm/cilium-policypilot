package hubble

import "time"

// Flow represents a single network flow observed by Hubble
type Flow struct {
	// Time information
	Time *time.Time `json:"time,omitempty"`

	// Source endpoint information
	Source *Endpoint `json:"source,omitempty"`

	// Destination endpoint information
	Destination *Endpoint `json:"destination,omitempty"`

	// Network layer information
	IP *IP `json:"ip,omitempty"`

	// Transport layer information
	L4 *Layer4 `json:"l4,omitempty"`

	// Flow verdict (ALLOWED, DENIED, etc.)
	Verdict string `json:"verdict,omitempty"`

	// Flow type (L3_L4, L7, etc.)
	Type *FlowType `json:"type,omitempty"`

	// Event type (PolicyVerdict, Trace, etc.)
	EventType *EventType `json:"event_type,omitempty"`
}

// Endpoint represents a network endpoint (pod, service, etc.)
type Endpoint struct {
	// Pod labels
	Labels []string `json:"labels,omitempty"`

	// Pod namespace
	Namespace string `json:"namespace,omitempty"`

	// Pod name
	PodName string `json:"pod_name,omitempty"`

	// Workload name
	Workloads []*Workload `json:"workloads,omitempty"`

	// Identity (security identity)
	Identity uint64 `json:"identity,omitempty"`
}

// Workload represents a Kubernetes workload
type Workload struct {
	// Workload name
	Name string `json:"name,omitempty"`

	// Workload kind (Deployment, StatefulSet, etc.)
	Kind string `json:"kind,omitempty"`
}

// IP represents IP layer information
type IP struct {
	// Source IP address
	Source string `json:"source,omitempty"`

	// Destination IP address
	Destination string `json:"destination,omitempty"`

	// IP version (4 or 6)
	IPVersion int `json:"ipVersion,omitempty"`
}

// Layer4 represents transport layer information
type Layer4 struct {
	// TCP information
	TCP *TCP `json:"TCP,omitempty"`

	// UDP information
	UDP *UDP `json:"UDP,omitempty"`
}

// TCP represents TCP protocol information
type TCP struct {
	// Source port
	SourcePort uint16 `json:"source_port,omitempty"`

	// Destination port
	DestinationPort uint16 `json:"destination_port,omitempty"`
}

// UDP represents UDP protocol information
type UDP struct {
	// Source port
	SourcePort uint16 `json:"source_port,omitempty"`

	// Destination port
	DestinationPort uint16 `json:"destination_port,omitempty"`
}

// FlowType represents the type of flow
type FlowType struct {
	Type int32 `json:"type,omitempty"`
}

// EventType represents the type of event
type EventType struct {
	Type int32 `json:"type,omitempty"`
}

// FlowCollection represents a collection of flows with metadata
type FlowCollection struct {
	Schema string  `json:"schema"`
	Flows  []*Flow `json:"flows"`
}

// ParsedFlow contains extracted metadata from a Flow for policy generation
type ParsedFlow struct {
	// Source pod labels (as map for easy lookup)
	SourceLabels map[string]string

	// Source namespace
	SourceNamespace string

	// Source pod name
	SourcePod string

	// Destination pod labels (as map for easy lookup)
	DestLabels map[string]string

	// Destination namespace
	DestNamespace string

	// Destination pod name
	DestPod string

	// Destination port
	DestPort uint16

	// Protocol (TCP, UDP, etc.)
	Protocol string

	// Direction (ingress/egress from destination perspective)
	Direction string

	// Verdict
	Verdict string
}

// ParseLabels converts a slice of label strings (format: "key=value") into a map
func ParseLabels(labelStrings []string) map[string]string {
	labels := make(map[string]string)
	for _, labelStr := range labelStrings {
		// Labels are typically in format "key=value"
		// Handle both "key=value" and just "key" formats
		found := false
		for i := 0; i < len(labelStr); i++ {
			if labelStr[i] == '=' {
				key := labelStr[:i]
				value := labelStr[i+1:]
				labels[key] = value
				found = true
				break
			}
		}
		// If no "=" found, treat entire string as key with empty value
		if !found && labelStr != "" {
			labels[labelStr] = ""
		}
	}
	return labels
}
