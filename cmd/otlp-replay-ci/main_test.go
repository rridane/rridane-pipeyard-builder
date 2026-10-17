package main

import (
	"slices"
	"testing"
)

// The command names are what the workflow writes, so they are the thing worth
// pinning.
func TestStagesAreWhatTheWorkflowCalls(t *testing.T) {
	var names []string
	for _, stage := range stages() {
		names = append(names, stage.Configure().Command())
	}

	want := []string{"guard", "secrets", "workflows",
		"lint", "typecheck", "architecture", "test", "lock-check", "audit"}
	if !slices.Equal(names, want) {
		t.Fatalf("stages are %v, want %v", names, want)
	}
}
