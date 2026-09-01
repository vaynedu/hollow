//go:build integration

package generator

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

	for _, command := range [][]string{
		{"make", "deps"},
		{"go", "test", "./..."},
		{"make", "build"},
	} {
		runCommand(t, target, command[0], command[1:]...)
	}
	assertGeneratedHealthEndpoint(t, target)
}

func assertGeneratedHealthEndpoint(t *testing.T, target string) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	_ = listener.Close()

	confPath := filepath.Join(target, "conf.yaml")
	conf := []byte("server:\n  host: " + address + "\n  shutdown_timeout: 2s\nlog:\n  level: error\n  output_mode: console\n")
	if err := os.WriteFile(confPath, conf, 0644); err != nil {
		t.Fatal(err)
	}

	command := exec.Command(filepath.Join(target, "bin", "lifelog-server"))
	command.Dir = target
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- command.Wait() }()
	waited := false
	defer func() {
		if waited {
			return
		}
		_ = command.Process.Signal(os.Interrupt)
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			_ = command.Process.Kill()
			<-done
		}
	}()

	client := &http.Client{Timeout: time.Second}
	var response *http.Response
	deadline := time.After(10 * time.Second)
	for response == nil {
		response, err = client.Get("http://" + address + "/v1/health")
		if err == nil {
			break
		}
		select {
		case processErr := <-done:
			waited = true
			t.Fatalf("generated server exited before health check: %v\n%s", processErr, output.String())
		case <-deadline:
			t.Fatalf("request generated health endpoint: %v\n%s", err, output.String())
		case <-time.After(20 * time.Millisecond):
		}
	}
	defer response.Body.Close()
	var envelope struct {
		Code int `json:"code"`
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || envelope.Code != 0 || envelope.Data.Status != "ok" {
		t.Fatalf("status=%d envelope=%+v", response.StatusCode, envelope)
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
