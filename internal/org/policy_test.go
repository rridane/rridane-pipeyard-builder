package org

import (
	"slices"
	"testing"
	"testing/fstest"

	"github.com/rridane/golang-pipeyard/x/stack/common"
)

// A project adds the shapes it produces; it does not replace the ones no
// repository may carry.
func TestBaselineAddsToTheOrganisationRules(t *testing.T) {
	guard := Baseline(DataShapes)[0].(*common.GuardStage)

	for _, want := range []string{"*.pem", "*.tfstate", "*.parquet"} {
		if !slices.Contains(guard.ForbiddenPaths, want) {
			t.Errorf("%s should be refused, paths are %v", want, guard.ForbiddenPaths)
		}
	}
	if slices.Contains(Baseline(nil)[0].(*common.GuardStage).ForbiddenPaths, "*.parquet") {
		t.Error("a project with no data shapes should not inherit another's")
	}
}

// The guard of a data project refuses the data, the key material and the
// denied words, and leaves the source alone.
func TestGuardRefusesDataKeysAndDeniedWords(t *testing.T) {
	guard := Baseline(DataShapes)[0].(*common.GuardStage)

	tree := fstest.MapFS{
		"src/app.py":                   {Data: []byte("print('ok')\n")},
		"data/spans.parquet":           {Data: []byte("PAR1")},
		"raw/minute=00/x.binpb.gz":     {Data: []byte{0x1f, 0x8b}},
		"warehouse/otel/metadata.json": {Data: []byte("{}")},
		"deploy/cluster.pem":           {Data: []byte("-----BEGIN\n")},
		"notes.md":                     {Data: []byte("deployed for acme corp\n")},
	}
	var files []string
	for name := range tree {
		files = append(files, name)
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
		"data/spans.parquet", "deploy/cluster.pem", "notes.md",
		"raw/minute=00/x.binpb.gz", "warehouse/otel/metadata.json",
	}
	if !slices.Equal(paths, want) {
		t.Fatalf("violations on %v, want %v", paths, want)
	}
}
