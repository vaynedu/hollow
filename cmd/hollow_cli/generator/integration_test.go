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
	handwrittenFiles := []string{
		"control/control.go",
		"control/health.go",
		"service/service.go",
		"service/health.go",
		"dao/dao.go",
		"model/model.go",
	}
	handwrittenBefore := make(map[string][]byte, len(handwrittenFiles))
	for _, relativePath := range handwrittenFiles {
		path := filepath.Join(target, relativePath)
		content := append(readFile(t, path), []byte("\n// keep handwritten code\n")...)
		if err := os.WriteFile(path, content, 0644); err != nil {
			t.Fatal(err)
		}
		handwrittenBefore[relativePath] = content
	}

	runCommand(t, target, "make", "proto")
	goModAfter := readFile(t, filepath.Join(target, "go.mod"))
	if !bytes.Equal(goModAfter, goModBefore) {
		t.Fatalf("make proto changed go.mod:\n%s", goModAfter)
	}
	for relativePath, before := range handwrittenBefore {
		after := readFile(t, filepath.Join(target, relativePath))
		if !bytes.Equal(after, before) {
			t.Errorf("make proto changed handwritten file %s", relativePath)
		}
	}
	writeGeneratedHealthTest(t, target)

	for _, command := range [][]string{
		{"make", "deps"},
		{"go", "test", "./..."},
		{"make", "build"},
	} {
		runCommand(t, target, command[0], command[1:]...)
	}
}

func writeGeneratedHealthTest(t *testing.T, target string) {
	t.Helper()
	const source = `package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vaynedu/hollow"
	"lifelog-server/router"
)

func TestGeneratedHealthEndpoint(t *testing.T) {
	app, err := hollow.NewApp(hollow.AppOption{ConfigPath: ".", ConfigName: "conf"})
	if err != nil {
		t.Fatal(err)
	}
	router.Register(app)

	recorder := httptest.NewRecorder()
	app.Engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/v1/lifelog-server/health", nil))
	var response struct {
		Code int ` + "`json:\"code\"`" + `
		Data struct {
			Status string ` + "`json:\"status\"`" + `
		} ` + "`json:\"data\"`" + `
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusOK || response.Code != 0 || response.Data.Status != "ok" {
		t.Fatalf("status=%d response=%+v body=%s", recorder.Code, response, recorder.Body.String())
	}
}
`
	if err := os.WriteFile(filepath.Join(target, "health_integration_test.go"), []byte(source), 0644); err != nil {
		t.Fatal(err)
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
