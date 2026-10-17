package main

import (
	"slices"
	"testing"
	"testing/fstest"

	"github.com/rridane/golang-pipeyard/x/stack/common"
)

func TestStagesAreTheOrganisationChecks(t *testing.T) {
	var names []string
	for _, stage := range stages() {
		names = append(names, stage.Configure().Command())
	}
	want := []string{
		"guard", "secrets", "lint", "typecheck",
		"architecture", "test", "lock-check", "audit",
	}
	if !slices.Equal(names, want) {
		t.Fatalf("stages = %v, want %v", names, want)
	}
}

func TestGuardRefusesDataAndDeniedWords(t *testing.T) {
	guard := stages()[0].(*common.GuardStage)
	tree := fstest.MapFS{
		"src/app.py":                   {Data: []byte("print('ok')\n")},
		"data/spans.parquet":           {Data: []byte("PAR1")},
		"raw/minute=00/x.binpb.gz":     {Data: []byte{0x1f, 0x8b}},
		"warehouse/otel/metadata.json": {Data: []byte("{}")},
		"notes.md":                     {Data: []byte("deployed for acme corp\n")},
	}
	files := []string{
		"src/app.py", "data/spans.parquet", "raw/minute=00/x.binpb.gz",
		"warehouse/otel/metadata.json", "notes.md",
	}

	violations, err := guard.Inspect(tree, files, common.ForbiddenWords("acme"))
	if err != nil {
		t.Fatal(err)
	}

	var paths []string
	for _, v := range violations {
		paths = append(paths, v.Path)
	}
	slices.Sort(paths)
	want := []string{
		"data/spans.parquet", "notes.md",
		"raw/minute=00/x.binpb.gz", "warehouse/otel/metadata.json",
	}
	if !slices.Equal(paths, want) {
		t.Fatalf("violations on %v, want %v", paths, want)
	}
}
