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
	root.SetOut(&bytes.Buffer{})
	root.SetArgs([]string{"--help"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	var names []string
	for _, cmd := range root.Commands() {
		if cmd.Name() == "help" {
			continue
		}
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

func TestInitCommandRequiresHollowPathDuringReleaseGate(t *testing.T) {
	called := false
	originalInitProject := initProject
	initProject = func(string, generator.ProjectOptions) error {
		called = true
		return nil
	}
	t.Cleanup(func() { initProject = originalInitProject })

	root := newRootCommand()
	root.SetArgs([]string{"init", "lifelog-server"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected hollow-path error")
	}
	for _, fragment := range []string{
		"v1.0.0",
		"Startup/Shutdown/Run",
		"pkg/hecode",
		"--hollow-path",
		"本地 Hollow",
	} {
		if !strings.Contains(err.Error(), fragment) {
			t.Errorf("error=%q, want %q", err, fragment)
		}
	}
	if called {
		t.Fatal("generator called without --hollow-path")
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

func TestInitCommandUsesDefaultHollowVersion(t *testing.T) {
	var gotOptions generator.ProjectOptions
	originalInitProject := initProject
	initProject = func(_ string, options generator.ProjectOptions) error {
		gotOptions = options
		return nil
	}
	t.Cleanup(func() { initProject = originalInitProject })

	root := newRootCommand()
	root.SetArgs([]string{"init", "lifelog-server", "--hollow-path", "/repo/hollow"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if got, want := gotOptions.HollowVersion, "v1.0.0"; got != want {
		t.Fatalf("hollow version=%q, want %q", got, want)
	}
}
