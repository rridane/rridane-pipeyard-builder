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

// acceptedAdvisories are the ones this repository has assessed and holds not
// to apply. It carries them and the other binaries do not, because it is the
// only one whose own dependency graph is this builder's. The stage fails on an
// identifier that matches nothing, so this list cannot outlive its reasons.
var acceptedAdvisories = []string{
	// golang.org/x/crypto is required for nacl/sign, through BuildKit's
	// auth provider. GO-2026-5932 is about its openpgp package, which has
	// no fixed version and which nothing here imports:
	//
	//	$ go mod why golang.org/x/crypto/openpgp
	//	(main module does not need package golang.org/x/crypto/openpgp)
	"GO-2026-5932",
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
		&common.AuditStage{Accept: acceptedAdvisories},
		&golang.BinariesStage{},
		&golang.PublishBinariesStage{},
	)
}
