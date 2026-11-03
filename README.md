# rridane-pipeyard-builder

The pipelines of my public repositories, and a working example of a
[Pipeyard](https://github.com/rridane/golang-pipeyard) pipeline builder repo.

Each repository gets its own binary, built here from the Pipeyard stage
catalogue. The repositories hold no pipeline logic: their workflow downloads
the latest release of their binary, checks where it was built, and calls the
stages one by one. The CI is a product of its own, with its own releases.

[otlp-replay](https://github.com/rridane/otlp-replay) is a consumer repo.

## Layout

```
internal/org/          what every repository answers to
  policy.go              forbidden shapes, size limit, denylist, zizmor policy
  baseline.go            guard, secrets, workflows
  run.go                 injectors and observer, the same for every binary
cmd/
  otlp-replay-ci/        one binary per repository: the stages it runs
  hune-lab-ci/
  rridane-pipeyard-builder-ci/
```

A repository's binary is a list of stages:

```go
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
```

Adding a repository is adding a directory under `cmd/`.

## Stages

Every repository answers to the same three first:

- `guard` refuses the shapes no repository may carry (private keys, Terraform
  state, a kubeconfig) plus the ones its own binary adds, files over 1 MiB,
  and the words listed in the `PIPEYARD_DENYLIST` secret. A violation reports
  the word's position in the list, never the word.
- `secrets` scans the whole git history with gitleaks.
- `workflows` audits the GitHub Actions workflows with zizmor. Actions may be
  pinned by tag. The policy lives here and is the only place an audit can be
  exempted.

Then the repository's own checks. A Python project runs `lint` (ruff), `typecheck`
(pyright), `architecture` (import-linter), `test` (pytest), `lock-check` and
`audit` (OSV), with the tools at the versions its own lockfile pins. A Go
module runs `build`, `fmt`, `vet`, `test`, `lint` and `audit`. A Node project
runs the scripts of its own `package.json` — `check`, `test` — once `npm ci` has
restored its lockfile, and `audit`.

A repository that ships images builds them last, one `image` stage per
Dockerfile target. Without `--push` the stage only checks that the image builds,
which is what a pull request runs; on `main` the workflow adds `--push`, and the
image is published under the commit SHA. The registry and its credentials come
from the environment, never a flag: `PIPEYARD_REGISTRY_HOST`,
`PIPEYARD_REGISTRY_USER` and `PIPEYARD_REGISTRY_TOKEN`.

## Releases

On every push to `main`, the builder rebuilds every binary. The builds are
reproducible, so a binary whose digest matches its last release has not
changed, whatever was touched. The others are attested, then released under
their own tag with the patch raised: `otlp-replay-ci/v0.1.4`.

A repository always runs the latest release of its binary, and runs it only
once the attestation proves it was built by this repository's workflow, from
`main`:

```yaml
- name: Fetch the pipeline
  run: |
    tag=$(gh release list --repo rridane/rridane-pipeyard-builder --exclude-drafts --exclude-pre-releases \
      --limit 100 --json tagName --jq "[.[].tagName | select(startswith(\"$PIPELINE/v\"))][0] // empty")
    test -n "$tag" || { echo "no release of $PIPELINE"; exit 1; }
    gh release download "$tag" --repo rridane/rridane-pipeyard-builder --pattern "$PIPELINE" --dir bin
    gh attestation verify "bin/$PIPELINE" --repo rridane/rridane-pipeyard-builder \
      --signer-workflow rridane/rridane-pipeyard-builder/.github/workflows/builder.yml \
      --source-ref refs/heads/main
    chmod +x "bin/$PIPELINE"

- run: bin/otlp-replay-ci guard
- run: bin/otlp-replay-ci test
```

A bad release is rolled back by deleting it: every repository falls back to
the previous one on its next run.

Locally, with a BuildKit daemon listening on `tcp://127.0.0.1:1234`:

```
go build -o bin/ ./cmd/...
cd ../otlp-replay && ../rridane-pipeyard-builder/bin/otlp-replay-ci lint
```

## License

Apache 2.0
