// Package org holds what every repository of this organisation answers to,
// whatever it is written in, and how each of its pipelines is wired. A
// project's binary adds to this; nothing it declares removes from it.
package org

// ForbiddenPaths are the shapes that are wrong in any repository, whatever it
// does. A project that produces shapes of its own refuses those too, next to
// these.
var ForbiddenPaths = []string{
	"*.pem", "*.p12", "*.pfx", "id_rsa", "id_ed25519",
	"*.tfstate", "*.tfstate.*",
	"kubeconfig", "*.kubeconfig",
}

// MaxFileSize is above the largest lockfile these repositories hold. Anything
// bigger is either a build output or data, and neither belongs in git.
const MaxFileSize = 1 << 20

// DenylistEnv names the secret holding the words that must appear nowhere in a
// public repository. The words themselves are never written here, and a
// violation reports a word's position in the list rather than the word.
const DenylistEnv = "PIPEYARD_DENYLIST"

// WorkflowsPolicy is what every repository's workflows answer to. Actions are
// pinned by tag rather than by commit, and the release job's cache is trusted.
// It is the only place an audit can be exempted: the stage runs --no-ignores,
// so a comment left in a workflow does not apply.
const WorkflowsPolicy = `rules:
  unpinned-uses:
    config:
      policies:
        "*": ref-pin
  cache-poisoning:
    disable: true
`

// DataShapes are the forms a trace archive, a dataset and a lakehouse take. A
// repository that reads and writes them may never carry one.
var DataShapes = []string{
	"*.parquet", "*.avro", "*.orc", "*.arrow", "*.feather",
	"*.pb", "*.pb.gz", "*.binpb", "*.binpb.gz", "*.json.gz",
	"*.pkl", "*.pt", "*.safetensors",
	"warehouse/*", "warehouse/*/*",
}
