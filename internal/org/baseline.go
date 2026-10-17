package org

import (
	"slices"

	"github.com/rridane/golang-pipeyard/pi"
	"github.com/rridane/golang-pipeyard/x/stack/common"
)

// Baseline is what every repository runs before anything else: what must never
// be published, in the order that refuses soonest. It takes the shapes the
// project refuses on top of the organisation's, and nothing else: a project
// configures what it adds, never what it drops.
func Baseline(extraPaths []string) []pi.Stage {
	return []pi.Stage{
		&common.GuardStage{
			ForbiddenPaths:    slices.Concat(ForbiddenPaths, extraPaths),
			MaxFileSize:       MaxFileSize,
			ForbiddenWordsEnv: DenylistEnv,
		},
		&common.SecretsStage{},
		&common.WorkflowsStage{Policy: WorkflowsPolicy},
	}
}
