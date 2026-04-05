package controller

import "testing"

func TestRolloutEnabled(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		ann  map[string]string
		want bool
	}{
		{"nil", nil, false},
		{"empty", map[string]string{}, false},
		{"false", map[string]string{RolloutAnnotation: "false"}, false},
		{"true", map[string]string{RolloutAnnotation: "true"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := rolloutEnabled(tt.ann); got != tt.want {
				t.Fatalf("rolloutEnabled(%v) = %v, want %v", tt.ann, got, tt.want)
			}
		})
	}
}
