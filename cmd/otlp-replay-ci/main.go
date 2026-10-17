// Command otlp-replay-ci is the pipeline of github.com/rridane/otlp-replay.
package main

import (
	"github.com/rridane/golang-pipeyard/pi"
	"github.com/rridane/golang-pipeyard/x/stack/common"
	"github.com/rridane/golang-pipeyard/x/stack/python"
	"github.com/rridane/rridane-pipeyard-builder/internal/org"
)

func main() {
	org.Run("otlp-replay-ci", "rridane/otlp-replay", stages())
}

// stages are what the workflow calls, in order. otlp-replay reads and writes
// traces, so it refuses to carry any.
func stages() []pi.Stage {
	return append(org.Baseline(org.DataShapes),
		&python.LintStage{},
		&python.TypeCheckStage{},
		&python.ArchitectureStage{},
		&python.PytestStage{},
		&python.LockStage{},
		&common.AuditStage{},
	)
}
