//go:build integration

package generator

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratedProjectEndToEnd(t *testing.T) {
	target := filepath.Join(t.TempDir(), "lifelog-server")
	if err := InitProject(target, ProjectOptions{HollowPath: repositoryRoot(t)}); err != nil {
		t.Fatal(err)
	}

	goModBefore := readFile(t, filepath.Join(target, "go.mod"))
	runCommand(t, target, "make", "proto")
	goModAfter := readFile(t, filepath.Join(target, "go.mod"))
	if !bytes.Equal(goModAfter, goModBefore) {
		t.Fatalf("make proto changed go.mod:\n%s", goModAfter)
	}

	for _, command := range [][]string{
		{"make", "deps"},
		{"go", "test", "./..."},
		{"go", "build", "./..."},
	} {
		runCommand(t, target, command[0], command[1:]...)
	}
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return content
}

func repositoryRoot(t *testing.T) string {
	t.Helper()

	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs(filepath.Join(workingDirectory, "..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func runCommand(t *testing.T, directory, name string, args ...string) {
	t.Helper()

	command := exec.Command(name, args...)
	command.Dir = directory
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, output)
	}
}
