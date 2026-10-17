package main

import (
	"slices"
	"testing"
)

func TestStagesAreWhatTheWorkflowCalls(t *testing.T) {
	var names []string
	for _, stage := range stages() {
		names = append(names, stage.Configure().Command())
	}

	want := []string{"guard", "secrets", "workflows",
		"build", "fmt", "vet", "test", "lint", "audit", "binaries", "publish-binaries"}
	if !slices.Equal(names, want) {
		t.Fatalf("stages are %v, want %v", names, want)
	}
}
