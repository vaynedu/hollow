package main

import (
	"bytes"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/vaynedu/hollow/cmd/hollow_cli/generator"
)

func TestRootCommandExposesOnlyInit(t *testing.T) {
	root := newRootCommand()
	var names []string
	for _, cmd := range root.Commands() {
		names = append(names, cmd.Name())
	}

	if !slices.Equal(names, []string{"init"}) {
		t.Fatalf("commands=%v, want [init]", names)
	}
}

func TestRootCommandReportsVersion(t *testing.T) {
	root := newRootCommand()
	var output bytes.Buffer
	root.SetOut(&output)
	root.SetArgs([]string{"--version"})

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if got, want := strings.TrimSpace(output.String()), "hollow-cli version dev"; got != want {
		t.Fatalf("version output=%q, want %q", got, want)
	}
}

func TestInitCommandRequiresExactlyOneProjectArgument(t *testing.T) {
	for _, args := range [][]string{{"init"}, {"init", "first", "second"}} {
		root := newRootCommand()
		root.SetArgs(args)
		if err := root.Execute(); err == nil {
			t.Fatalf("args=%v: expected argument error", args)
		}
	}
}

func TestInitCommandPassesAllOptionsToGenerator(t *testing.T) {
	wantProject := "lifelog-server"
	wantOptions := generator.ProjectOptions{
		Module:        "example.com/lifelog",
		Service:       "DiaryService",
		HollowVersion: "v1.2.0",
		HollowPath:    "/repo/hollow",
	}
	var gotProject string
	var gotOptions generator.ProjectOptions
	originalInitProject := initProject
	initProject = func(project string, options generator.ProjectOptions) error {
		gotProject = project
		gotOptions = options
		return nil
	}
	t.Cleanup(func() { initProject = originalInitProject })

	root := newRootCommand()
	root.SetArgs([]string{
		"init", wantProject,
		"--module", wantOptions.Module,
		"--service", wantOptions.Service,
		"--hollow-version", wantOptions.HollowVersion,
		"--hollow-path", wantOptions.HollowPath,
	})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if gotProject != wantProject {
		t.Fatalf("project=%q, want %q", gotProject, wantProject)
	}
	if !reflect.DeepEqual(gotOptions, wantOptions) {
		t.Fatalf("options=%+v, want %+v", gotOptions, wantOptions)
	}
}
