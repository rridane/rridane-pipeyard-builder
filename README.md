# rr-ci

The pipeline of my Python repositories, written with
[Pipeyard](https://github.com/rridane/golang-pipeyard).

A repository holds no pipeline logic. Its workflow downloads a pinned release
of `rr-ci`, checks where it was built, and calls the stages one by one. What a
stage needs to know about the project comes from the project itself:
`pyproject.toml` and `uv.lock`.

## Stages

In the order a workflow calls them:

- `guard` refuses data files (Parquet, Arrow, raw OTLP, model weights), files
  over 1 MiB and the words listed in the `RR_CI_DENYLIST` secret. A violation
  reports the word's position in the list, never the word.
- `secrets` scans the whole git history with gitleaks.
- `lint` runs ruff, checks and formatting.
- `typecheck` runs pyright.
- `architecture` runs the import-linter contracts.
- `test` runs pytest.
- `lock-check` checks that `uv.lock` still matches `pyproject.toml`.
- `audit` looks the locked dependencies up in the OSV database.

The Python tools run at the versions the project's own lockfile pins, in a
container, so a laptop and the CI give the same answer.

## Use it in a repository

```yaml
- uses: actions/checkout@v4
  with:
    fetch-depth: 0

- name: Fetch rr-ci
  run: |
    gh release download v0.1.0 --repo rridane/rr-ci --pattern rr-ci --dir bin
    gh attestation verify bin/rr-ci --repo rridane/rr-ci
    chmod +x bin/rr-ci
  env:
    GH_TOKEN: ${{ github.token }}

- name: Start BuildKit
  run: |
    docker run -d --name buildkit --privileged -p 1234:1234 moby/buildkit:v0.34.0 --addr tcp://0.0.0.0:1234
    timeout 60 sh -c 'until docker logs buildkit 2>&1 | grep -q "running server"; do sleep 1; done'

- run: bin/rr-ci guard
  env:
    RR_CI_DENYLIST: ${{ secrets.RR_CI_DENYLIST }}
- run: bin/rr-ci test
```

Locally, with a BuildKit daemon listening on `tcp://127.0.0.1:1234`:

```
go build -o bin/rr-ci .
cd ../my-repository && ../rr-ci/bin/rr-ci lint
```

## License

Apache 2.0
