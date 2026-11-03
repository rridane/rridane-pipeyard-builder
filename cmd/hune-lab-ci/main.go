// Command hune-lab-ci is the pipeline of github.com/rridane/hune-lab.
package main

import (
	"github.com/rridane/golang-pipeyard/pi"
	"github.com/rridane/golang-pipeyard/x/stack/common"
	"github.com/rridane/golang-pipeyard/x/stack/node"
	"github.com/rridane/rridane-pipeyard-builder/internal/org"
)

func main() {
	org.Run("hune-lab-ci", "rridane/hune-lab", stages())
}

// catalogue is where the workflow puts the articles of the catalogue
// repository. Git ignores it: the site follows that repository at run time,
// and the copy only serves the parity check and the snapshot the images ship.
var catalogue = []string{"articles"}

// stages are what the workflow calls, in order. The images come last, after
// everything that could refuse the change; the workflow pushes them on main
// only.
func stages() []pi.Stage {
	return append(org.Baseline(nil),
		&node.ScriptStage{Script: "check"},
		&node.ScriptStage{Script: "test", Include: catalogue},
		&node.ScriptStage{Script: "test:parity", Command: "parity", Include: catalogue},
		&common.AuditStage{},
		&common.ImageStage{Repository: "ghcr.io/rridane/hune-lab", Target: "runtime", Include: catalogue},
		&common.ImageStage{
			Repository: "ghcr.io/rridane/hune-lab-web",
			Target:     "web",
			Command:    "image-web",
			Include:    catalogue,
		},
	)
}
