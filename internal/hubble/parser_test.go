package hubble

import (
	"testing"
)

func TestParseLabels(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected map[string]string
	}{
		{
			name:     "empty labels",
			input:    []string{},
			expected: map[string]string{},
		},
		{
			name:  "single label",
			input: []string{"app=frontend"},
			expected: map[string]string{
				"app": "frontend",
			},
		},
		{
			name:  "multiple labels",
			input: []string{"k8s:app=frontend", "k8s:version=v1", "k8s:io.kubernetes.pod.namespace=default"},
			expected: map[string]string{
				"k8s:app":                         "frontend",
				"k8s:version":                     "v1",
				"k8s:io.kubernetes.pod.namespace": "default",
			},
		},
		{
			name:  "labels without values",
			input: []string{"app"},
			expected: map[string]string{
				"app": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseLabels(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("ParseLabels() length = %d, want %d", len(result), len(tt.expected))
			}

			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("ParseLabels() [%s] = %s, want %s", k, result[k], v)
				}
			}
		})
	}
}

func TestParseFlow(t *testing.T) {
	tests := []struct {
		name     string
		flow     *Flow
		wantErr  bool
		validate func(*testing.T, *ParsedFlow)
	}{
		{
			name:    "nil flow",
			flow:    nil,
			wantErr: true,
		},
		{
			name: "valid TCP flow",
			flow: &Flow{
				Source: &Endpoint{
					Labels:    []string{"k8s:app=frontend"},
					Namespace: "default",
					PodName:   "frontend-pod",
				},
				Destination: &Endpoint{
					Labels:    []string{"k8s:app=catalog"},
					Namespace: "default",
					PodName:   "catalog-pod",
				},
				L4: &Layer4{
					TCP: &TCP{
						DestinationPort: 8080,
					},
				},
				Verdict: "ALLOWED",
			},
			wantErr: false,
			validate: func(t *testing.T, pf *ParsedFlow) {
				if pf.Protocol != "TCP" {
					t.Errorf("Protocol = %s, want TCP", pf.Protocol)
				}
				if pf.DestPort != 8080 {
					t.Errorf("DestPort = %d, want 8080", pf.DestPort)
				}
				if pf.SourceNamespace != "default" {
					t.Errorf("SourceNamespace = %s, want default", pf.SourceNamespace)
				}
				if pf.SourceLabels["k8s:app"] != "frontend" {
					t.Errorf("SourceLabels[k8s:app] = %s, want frontend", pf.SourceLabels["k8s:app"])
				}
			},
		},
		{
			name: "valid UDP flow",
			flow: &Flow{
				Source: &Endpoint{
					Labels:    []string{"k8s:app=dns"},
					Namespace: "kube-system",
				},
				Destination: &Endpoint{
					Labels:    []string{"k8s:app=client"},
					Namespace: "default",
				},
				L4: &Layer4{
					UDP: &UDP{
						DestinationPort: 53,
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, pf *ParsedFlow) {
				if pf.Protocol != "UDP" {
					t.Errorf("Protocol = %s, want UDP", pf.Protocol)
				}
				if pf.DestPort != 53 {
					t.Errorf("DestPort = %d, want 53", pf.DestPort)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseFlow(tt.flow)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}
