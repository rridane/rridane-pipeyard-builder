// Command rridane-pipeyard-builder-ci is the pipeline of this repository: it
// checks the Go, then builds every binary under cmd/ and releases those that
// changed.
package main

import (
	"github.com/rridane/golang-pipeyard/pi"
	"github.com/rridane/golang-pipeyard/x/stack/common"
	"github.com/rridane/golang-pipeyard/x/stack/golang"
	"github.com/rridane/rridane-pipeyard-builder/internal/org"
)

func main() {
	org.Run("rridane-pipeyard-builder-ci", "rridane/rridane-pipeyard-builder", stages())
}

// stages are what the workflow calls, in order. The release runs last, after
// everything that could refuse the change.
func stages() []pi.Stage {
	return append(org.Baseline(nil),
		&golang.BuildStage{},
		&golang.FormatStage{},
		&golang.VetStage{},
		&golang.TestStage{},
		&golang.LintStage{},
		&common.AuditStage{},
		&golang.BinariesStage{},
		&golang.PublishBinariesStage{},
	)
}
