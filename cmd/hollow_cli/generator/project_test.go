package generator

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestNewProjectConfigDerivesDefaults(t *testing.T) {
	cfg, err := newProjectConfig("/tmp/lifelog-server", ProjectOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if cfg.ProjectName != "lifelog-server" || cfg.ModuleName != "lifelog-server" ||
		cfg.ServiceName != "LifelogService" || cfg.ProtoName != "lifelog" ||
		cfg.HollowVersion != "v1.0.0" {
		t.Fatalf("config=%+v", cfg)
	}
}

func TestNewProjectConfigUsesOverrides(t *testing.T) {
	cfg, err := newProjectConfig("lifelog-server", ProjectOptions{
		Module:        "example.com/lifelog",
		Service:       "DiaryService",
		HollowVersion: "v1.2.0",
		HollowPath:    "/repo/hollow",
	})
	if err != nil {
		t.Fatal(err)
	}

	if cfg.ModuleName != "example.com/lifelog" || cfg.ServiceName != "DiaryService" ||
		cfg.HollowVersion != "v1.2.0" || cfg.HollowPath != "/repo/hollow" {
		t.Fatalf("config=%+v", cfg)
	}
}

func TestDeriveNames(t *testing.T) {
	tests := []struct {
		name        string
		wantService string
		wantProto   string
	}{
		{name: "lifelog-server", wantService: "LifelogService", wantProto: "lifelog"},
		{name: "lifelog_server", wantService: "LifelogService", wantProto: "lifelog"},
		{name: "order-center-server", wantService: "OrderCenterService", wantProto: "order_center"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deriveServiceName(tt.name); got != tt.wantService {
				t.Errorf("deriveServiceName(%q)=%q, want %q", tt.name, got, tt.wantService)
			}
			if got := deriveProtoName(tt.name); got != tt.wantProto {
				t.Errorf("deriveProtoName(%q)=%q, want %q", tt.name, got, tt.wantProto)
			}
		})
	}
}

func TestInitProjectRejectsExistingTarget(t *testing.T) {
	target := t.TempDir()
	err := InitProject(target, ProjectOptions{})
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("error=%v", err)
	}
}

func TestInitProjectCleansTemporaryDirectoryOnRenderFailure(t *testing.T) {
	parent := t.TempDir()
	target := filepath.Join(parent, "lifelog-server")
	keep := filepath.Join(parent, ".lifelog-server-keep")
	if err := os.Mkdir(keep, 0755); err != nil {
		t.Fatal(err)
	}

	renderErr := errors.New("render failed")
	calls := 0
	originalWriter := projectTemplateWriter
	projectTemplateWriter = func(path, templateName string, config ProjectConfig) error {
		calls++
		if calls == 2 {
			return renderErr
		}
		return writeTemplate(path, templateName, config)
	}
	t.Cleanup(func() { projectTemplateWriter = originalWriter })

	err := InitProject(target, ProjectOptions{})
	if !errors.Is(err, renderErr) {
		t.Fatalf("error=%v, want %v", err, renderErr)
	}
	if _, err := os.Stat(target); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("target stat error=%v, want not exist", err)
	}
	if _, err := os.Stat(keep); err != nil {
		t.Fatalf("unrelated directory removed: %v", err)
	}

	entries, err := os.ReadDir(parent)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != filepath.Base(keep) {
		t.Fatalf("parent entries=%v, want only %q", entries, filepath.Base(keep))
	}
}

func TestInitProjectGeneratesGoldenTree(t *testing.T) {
	target := filepath.Join(t.TempDir(), "lifelog-server")
	if err := InitProject(target, ProjectOptions{}); err != nil {
		t.Fatal(err)
	}

	want := []string{
		".gitignore",
		"Makefile",
		"README.md",
		"conf.yaml",
		"config/config.go",
		"go.mod",
		"main.go",
		"proto/lifelog.proto",
		"router/router.go",
		"service/service.go",
	}
	assertExactTree(t, target, want)
}

func TestInitProjectDefaultGoModOmitsReplace(t *testing.T) {
	target := filepath.Join(t.TempDir(), "lifelog-server")
	if err := InitProject(target, ProjectOptions{}); err != nil {
		t.Fatal(err)
	}

	content := readProjectFile(t, target, "go.mod")
	if strings.Contains(content, "replace github.com/vaynedu/hollow") {
		t.Fatalf("go.mod contains unexpected replace:\n%s", content)
	}
}

func TestInitProjectHollowPathAddsExactReplace(t *testing.T) {
	target := filepath.Join(t.TempDir(), "lifelog-server")
	if err := InitProject(target, ProjectOptions{HollowPath: "/repo/hollow"}); err != nil {
		t.Fatal(err)
	}

	content := readProjectFile(t, target, "go.mod")
	want := "replace github.com/vaynedu/hollow => /repo/hollow\n"
	if strings.Count(content, want) != 1 {
		t.Fatalf("go.mod exact replace count=%d, want 1:\n%s", strings.Count(content, want), content)
	}
}

func TestWriteTemplateErrorIncludesTargetPath(t *testing.T) {
	cfg, err := newProjectConfig("lifelog-server", ProjectOptions{})
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(t.TempDir(), "missing", "go.mod")

	err = writeTemplate(target, "go.mod.tmpl", cfg)
	if err == nil || !strings.Contains(err.Error(), target) {
		t.Fatalf("error=%v, want target path %q", err, target)
	}
}

func TestInitProjectRendersStableStandaloneTemplates(t *testing.T) {
	target := filepath.Join(t.TempDir(), "lifelog-server")
	if err := InitProject(target, ProjectOptions{}); err != nil {
		t.Fatal(err)
	}

	mainFile := readProjectFile(t, target, "main.go")
	assertContainsAll(t, "main.go", mainFile,
		"hollow.NewApp(",
		"app.Startup(config.Startup)",
		"app.Shutdown(config.Shutdown)",
		"router.Register(app)",
		"return app.Run()",
	)
	if strings.Contains(mainFile, "app.Start()") || strings.Contains(mainFile, "app.End()") {
		t.Fatalf("main.go still uses removed lifecycle methods:\n%s", mainFile)
	}

	serviceFile := readProjectFile(t, target, "service/service.go")
	assertContainsAll(t, "service/service.go", serviceFile,
		"proto.UnimplementedLifelogServiceService",
		"func Get() *Service",
		"return instance",
	)

	routerFile := readProjectFile(t, target, "router/router.go")
	assertContainsAll(t, "router/router.go", routerFile,
		"proto.RegisterLifelogServiceGinRouter(",
		"app.Engine",
		"service.Get(),",
	)

	makefile := readProjectFile(t, target, "Makefile")
	assertContainsAll(t, "Makefile", makefile,
		"PROTO_INCLUDE ?= /usr/local/include",
		"protoc proto/*.proto \\",
		"-I . -I $(PROTO_INCLUDE) \\",
		"--go_out=. --go_opt=paths=source_relative \\",
		"--myhttp_out=. --myhttp_opt=paths=source_relative",
	)
	for _, forbidden := range []string{
		"hollow-cli",
		"go mod tidy",
		"openapi",
		"grpc-gateway@",
		"$(shell go env GOPATH)",
	} {
		if strings.Contains(makefile, forbidden) {
			t.Fatalf("Makefile contains forbidden %q:\n%s", forbidden, makefile)
		}
	}
}

func assertExactTree(t *testing.T, root string, want []string) {
	t.Helper()
	var got []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		got = append(got, filepath.ToSlash(relative))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("tree=%v, want %v", got, want)
	}
}

func readProjectFile(t *testing.T, root, relativePath string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(root, relativePath))
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func assertContainsAll(t *testing.T, name, content string, want ...string) {
	t.Helper()
	for _, fragment := range want {
		if !strings.Contains(content, fragment) {
			t.Errorf("%s missing %q:\n%s", name, fragment, content)
		}
	}
}
