package synth

import (
	"testing"

	"github.com/prabhakaran-jm/cilium-policypilot/internal/hubble"
)

func TestSynthesizePolicies(t *testing.T) {
	tests := []struct {
		name     string
		flows    []*hubble.ParsedFlow
		wantErr  bool
		validate func(*testing.T, []*Policy)
	}{
		{
			name:    "empty flows",
			flows:   []*hubble.ParsedFlow{},
			wantErr: true,
		},
		{
			name: "single flow",
			flows: []*hubble.ParsedFlow{
				{
					SourceLabels:    map[string]string{"k8s:app": "frontend"},
					SourceNamespace: "default",
					DestLabels:      map[string]string{"k8s:app": "catalog"},
					DestNamespace:   "default",
					DestPort:        8080,
					Protocol:        "TCP",
				},
			},
			wantErr: false,
			validate: func(t *testing.T, policies []*Policy) {
				if len(policies) != 1 {
					t.Errorf("Expected 1 policy, got %d", len(policies))
					return
				}
				policy := policies[0]
				if policy.Metadata.Name != "catalog-policy" {
					t.Errorf("Expected policy name 'catalog-policy', got '%s'", policy.Metadata.Name)
				}
				if len(policy.Spec.Ingress) != 1 {
					t.Errorf("Expected 1 ingress rule, got %d", len(policy.Spec.Ingress))
				}
			},
		},
		{
			name: "multiple flows to same destination",
			flows: []*hubble.ParsedFlow{
				{
					SourceLabels:    map[string]string{"k8s:app": "frontend"},
					SourceNamespace: "default",
					DestLabels:      map[string]string{"k8s:app": "catalog"},
					DestNamespace:   "default",
					DestPort:        8080,
					Protocol:        "TCP",
				},
				{
					SourceLabels:    map[string]string{"k8s:app": "frontend"},
					SourceNamespace: "default",
					DestLabels:      map[string]string{"k8s:app": "catalog"},
					DestNamespace:   "default",
					DestPort:        8081,
					Protocol:        "TCP",
				},
			},
			wantErr: false,
			validate: func(t *testing.T, policies []*Policy) {
				if len(policies) != 1 {
					t.Errorf("Expected 1 policy, got %d", len(policies))
					return
				}
				policy := policies[0]
				if len(policy.Spec.Ingress[0].ToPorts[0].Ports) != 2 {
					t.Errorf("Expected 2 ports, got %d", len(policy.Spec.Ingress[0].ToPorts[0].Ports))
				}
			},
		},
		{
			name: "flows without destination labels",
			flows: []*hubble.ParsedFlow{
				{
					SourceLabels:    map[string]string{"k8s:app": "frontend"},
					SourceNamespace: "default",
					DestLabels:      map[string]string{},
					DestNamespace:   "default",
					DestPort:        8080,
					Protocol:        "TCP",
				},
			},
			wantErr: false,
			validate: func(t *testing.T, policies []*Policy) {
				// Should not generate policy for flows without destination labels
				if len(policies) != 0 {
					t.Errorf("Expected 0 policies for flows without destination labels, got %d", len(policies))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policies, err := SynthesizePolicies(tt.flows)

			if (err != nil) != tt.wantErr {
				t.Errorf("SynthesizePolicies() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, policies)
			}
		})
	}
}

func TestGeneratePolicyName(t *testing.T) {
	tests := []struct {
		name     string
		labels   map[string]string
		expected string
	}{
		{
			name:     "app label",
			labels:   map[string]string{"k8s:app": "frontend"},
			expected: "frontend-policy",
		},
		{
			name:     "k8s:app label",
			labels:   map[string]string{"k8s:app": "catalog"},
			expected: "catalog-policy",
		},
		{
			name:     "name label",
			labels:   map[string]string{"name": "myapp"},
			expected: "myapp-policy",
		},
		{
			name:     "no preferred labels",
			labels:   map[string]string{"version": "v1"},
			expected: "v1-policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generatePolicyName(tt.labels)
			if result != tt.expected {
				t.Errorf("generatePolicyName() = %v, want %v", result, tt.expected)
			}
		})
	}
}
