package main

import (
	"github.com/rridane/golang-pipeyard/pi"
	"github.com/rridane/golang-pipeyard/x/stack/common"
	"github.com/rridane/golang-pipeyard/x/stack/python"
)

// forbiddenPaths are the shapes a trace archive, a dataset and a lakehouse
// take. These repositories read and write them; none of them may carry one.
var forbiddenPaths = []string{
	"*.parquet", "*.avro", "*.orc", "*.arrow", "*.feather",
	"*.pb", "*.pb.gz", "*.binpb", "*.binpb.gz", "*.json.gz",
	"*.pkl", "*.pt", "*.safetensors",
	"warehouse/*", "warehouse/*/*",
}

// maxFileSize is above the largest lockfile any of these repositories holds.
const maxFileSize = 1 << 20

// denylistEnv holds the words that must appear nowhere in a public repository.
// They come from a CI secret, so they are never written here, and a violation
// reports a word's position in the list rather than the word.
const denylistEnv = "RR_CI_DENYLIST"

// stages are the organisation's checks, in the order a workflow calls them:
// what must never be published first, the code next, the dependencies last.
func stages() []pi.Stage {
	return []pi.Stage{
		&common.GuardStage{
			ForbiddenPaths:    forbiddenPaths,
			MaxFileSize:       maxFileSize,
			ForbiddenWordsEnv: denylistEnv,
		},
		&common.SecretsStage{},
		&python.LintStage{},
		&python.TypeCheckStage{},
		&python.ArchitectureStage{},
		&python.PytestStage{},
		&python.LockStage{},
		&common.AuditStage{},
	}
}
