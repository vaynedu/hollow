package generator

import (
	"bytes"
	"errors"
	"go/format"
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
		cfg.EnvPrefix != "LIFELOG" || cfg.HollowVersion != "v1.0.0" {
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
	originalWriter := projectTemplateWriter
	projectTemplateWriter = func(path, templateName string, config ProjectConfig) error {
		if templateName == "README.md.tmpl" {
			return renderErr
		}
		return writeTemplate(path, templateName, config)
	}
	t.Cleanup(func() { projectTemplateWriter = originalWriter })

	err := InitProject(target, ProjectOptions{})
	if !errors.Is(err, renderErr) {
		t.Fatalf("error=%v, want %v", err, renderErr)
	}
	finalFile := filepath.Join(target, "README.md")
	if !strings.Contains(err.Error(), finalFile) {
		t.Fatalf("error=%v, want final file path %q", err, finalFile)
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

func TestInitProjectNeverReplacesTargetCreatedBeforePublish(t *testing.T) {
	tests := []struct {
		name         string
		withSentinel bool
	}{
		{name: "empty directory"},
		{name: "directory with sentinel", withSentinel: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := t.TempDir()
			target := filepath.Join(parent, "lifelog-server")
			var hookErr error
			created := false
			originalWriter := projectTemplateWriter
			projectTemplateWriter = func(path, templateName string, config ProjectConfig) error {
				if err := writeTemplate(path, templateName, config); err != nil {
					return err
				}
				if created {
					return nil
				}
				created = true
				hookErr = os.Mkdir(target, 0755)
				if hookErr == nil && tt.withSentinel {
					hookErr = os.WriteFile(filepath.Join(target, "sentinel"), []byte("keep"), 0644)
				}
				return hookErr
			}
			defer func() { projectTemplateWriter = originalWriter }()

			err := InitProject(target, ProjectOptions{})
			if hookErr != nil {
				t.Fatalf("publish hook: %v", hookErr)
			}
			if err == nil || !strings.Contains(err.Error(), "already exists") {
				t.Fatalf("error=%v, want already exists", err)
			}

			entries, err := os.ReadDir(target)
			if err != nil {
				t.Fatal(err)
			}
			if tt.withSentinel {
				if len(entries) != 1 || entries[0].Name() != "sentinel" {
					t.Fatalf("target entries=%v, want sentinel", entries)
				}
				content, err := os.ReadFile(filepath.Join(target, "sentinel"))
				if err != nil || string(content) != "keep" {
					t.Fatalf("sentinel content=%q error=%v", content, err)
				}
			} else if len(entries) != 0 {
				t.Fatalf("target entries=%v, want empty", entries)
			}

			parentEntries, err := os.ReadDir(parent)
			if err != nil {
				t.Fatal(err)
			}
			if len(parentEntries) != 1 || parentEntries[0].Name() != filepath.Base(target) {
				t.Fatalf("parent entries=%v, want only target", parentEntries)
			}
		})
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
		"config/config_test.go",
		"control/control.go",
		"control/health.go",
		"control/health_test.go",
		"dao/dao.go",
		"database/database.go",
		"database/mysql.go",
		"database/redis.go",
		"go.mod",
		"main.go",
		"model/model.go",
		"model/user.go",
		"model/user_test.go",
		"proto/lifelog.proto",
		"router/router.go",
		"service/health.go",
		"service/health_test.go",
		"service/service.go",
	}
	assertExactTree(t, target, want)
}

func TestInitProjectRendersGofmtStableGoFiles(t *testing.T) {
	target := filepath.Join(t.TempDir(), "lifelog-server")
	if err := InitProject(target, ProjectOptions{}); err != nil {
		t.Fatal(err)
	}

	err := filepath.WalkDir(target, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}

		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		formatted, err := format.Source(source)
		if err != nil {
			t.Errorf("format %s: %v", path, err)
			return nil
		}
		if !bytes.Equal(source, formatted) {
			relative, err := filepath.Rel(target, path)
			if err != nil {
				return err
			}
			t.Errorf("%s is not gofmt stable:\n--- rendered\n%s\n--- formatted\n%s", relative, source, formatted)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
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
		"EnvPrefix:  config.EnvPrefix",
		"config.Startup",
		"database.InitMySQL",
		"database.InitRedis",
		"app.Shutdown(database.Shutdown)",
		"router.Register(app)",
		"return app.Run()",
	)
	if strings.Contains(mainFile, "app.Start()") || strings.Contains(mainFile, "app.End()") {
		t.Fatalf("main.go still uses removed lifecycle methods:\n%s", mainFile)
	}

	controlFile := readProjectFile(t, target, "control/control.go")
	assertContainsAll(t, "control/control.go", controlFile,
		"proto.UnimplementedLifelogServiceService",
		"func Get() *Control",
		"return instance",
	)
	healthControlFile := readProjectFile(t, target, "control/health.go")
	assertContainsAll(t, "control/health.go", healthControlFile,
		"hlog.FromContext(ctx)",
		"func (Control) Health(",
		"service.Health(ctx)",
		"&proto.HealthResponse{Status: status}",
	)
	configFile := readProjectFile(t, target, "config/config.go")
	assertContainsAll(t, "config/config.go", configFile,
		`const EnvPrefix = "LIFELOG"`,
		"type Config struct",
		"type DatabaseConfig struct",
		"func Startup() error",
		"func Get() *Config",
		"hconfig.LoadWithEnv(path, name, EnvPrefix, cfg)",
		"func (c *Config) Validate() error",
	)
	databaseFile := readProjectFile(t, target, "database/database.go")
	assertContainsAll(t, "database/database.go", databaseFile,
		"func Shutdown(ctx context.Context) error",
		"errors.Join",
	)
	mysqlFile := readProjectFile(t, target, "database/mysql.go")
	assertContainsAll(t, "database/mysql.go", mysqlFile,
		"func InitMySQL() error",
		"hgorm.NewMySQL",
		"func MySQL() *gorm.DB",
	)
	redisFile := readProjectFile(t, target, "database/redis.go")
	assertContainsAll(t, "database/redis.go", redisFile,
		"func InitRedis() error",
		"hredis.NewClient",
		"func Redis() *redis.Client",
	)
	confFile := readProjectFile(t, target, "conf.yaml")
	assertContainsAll(t, "conf.yaml", confFile,
		"database:",
		"mysql:",
		"redis:",
		"enabled: false",
	)
	goMod := readProjectFile(t, target, "go.mod")
	assertContainsAll(t, "go.mod", goMod,
		"github.com/smartystreets/goconvey",
		"github.com/redis/go-redis/v9",
		"gorm.io/gorm",
	)
	for _, file := range []string{
		"config/config_test.go",
		"control/health_test.go",
		"service/health_test.go",
	} {
		content := readProjectFile(t, target, file)
		assertContainsAll(t, file, content, "Convey(", "So(")
	}
	userModel := readProjectFile(t, target, "model/user.go")
	assertContainsAll(t, "model/user.go", userModel,
		`const TableNameUser = "users"`,
		"UserStatusNormal",
		"UserStatusDisabled",
		"type User struct",
		`gorm:"column:id;primary_key;AUTO_INCREMENT;comment:用户ID" json:"id"`,
		`gorm:"column:create_time;default:CURRENT_TIMESTAMP(3);comment:创建时间" json:"create_time"`,
		"func (u *User) TableName() string",
	)
	userModelTest := readProjectFile(t, target, "model/user_test.go")
	assertContainsAll(t, "model/user_test.go", userModelTest,
		"Convey(",
		"So(",
		"TableNameUser",
	)
	healthServiceFile := readProjectFile(t, target, "service/health.go")
	assertContainsAll(t, "service/health.go", healthServiceFile,
		"func Health(ctx context.Context) (string, error)",
		`return "ok", nil`,
	)
	protoFile := readProjectFile(t, target, "proto/lifelog.proto")
	assertContainsAll(t, "proto/lifelog.proto", protoFile,
		"message HealthRequest {}",
		"message HealthResponse {",
		"rpc Health(HealthRequest) returns (HealthResponse)",
		`get: "/v1/lifelog-server/health"`,
	)
	for _, forbidden := range []string{"PingRequest", "PingResponse", "rpc Ping", `get: "/v1/ping"`} {
		if strings.Contains(protoFile, forbidden) {
			t.Errorf("proto/lifelog.proto contains forbidden %q:\n%s", forbidden, protoFile)
		}
	}

	for _, file := range []string{"service/service.go", "service/health.go", "dao/dao.go", "model/model.go"} {
		content := readProjectFile(t, target, file)
		for _, forbidden := range []string{"/proto", "gin-gonic", "gorm.io"} {
			if strings.Contains(content, forbidden) {
				t.Errorf("%s contains forbidden %q:\n%s", file, forbidden, content)
			}
		}
	}

	routerFile := readProjectFile(t, target, "router/router.go")
	assertContainsAll(t, "router/router.go", routerFile,
		"proto.RegisterLifelogServiceGinRouter(",
		"app.Engine",
		"control.Get(),",
	)
	if strings.Contains(routerFile, "service.Get()") {
		t.Fatalf("router/router.go still registers service:\n%s", routerFile)
	}

	makefile := readProjectFile(t, target, "Makefile")
	assertContainsAll(t, "Makefile", makefile,
		"PROTO_INCLUDE ?= /usr/local/include",
		"protoc proto/*.proto \\",
		"-I . -I $(PROTO_INCLUDE) \\",
		"--go_out=. --go_opt=paths=source_relative \\",
		"--myhttp_out=. --myhttp_opt=paths=source_relative",
		"deps:\n\tgo mod tidy",
	)
	for _, forbidden := range []string{
		"hollow-cli",
		"openapi",
		"grpc-gateway@",
		"$(shell go env GOPATH)",
	} {
		if strings.Contains(makefile, forbidden) {
			t.Fatalf("Makefile contains forbidden %q:\n%s", forbidden, makefile)
		}
	}
}

func TestInitProjectREADMEExplainsUnreleasedWorkflow(t *testing.T) {
	target := filepath.Join(t.TempDir(), "lifelog-server")
	if err := InitProject(target, ProjectOptions{}); err != nil {
		t.Fatal(err)
	}

	readme := readProjectFile(t, target, "README.md")
	assertContainsAll(t, "README.md", readme,
		"Go 1.25+",
		"Protocol Buffers 编译器",
		"git clone https://github.com/googleapis/googleapis.git /absolute/path/to/googleapis",
		"go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.1\n",
		"(cd /path/to/protoc-gen-myhttp && go install .)",
		"make proto PROTO_INCLUDE=/absolute/path/to/googleapis",
		"/absolute/path/to/googleapis/google/api/annotations.proto",
		"hollow-cli init lifelog-server --hollow-path /path/to/hollow",
		"最新的生命周期、错误码、配置、日志和数据库 API",
		"发布包含新 API 的 Hollow 版本",
		"`DefaultHollowVersion` 更新后",
		"才可省略 `--hollow-path`",
		"现有 `v1.0.0` 标签不得移动或覆盖",
		"hlog.FromContext(ctx)",
		`hlog.FromContext(ctx).Named("HealthControl.Health")`,
		`logger.Info("health checked")`,
		"MySQL 和 Redis 默认关闭",
		"GoConvey",
		"model/user.go",
	)
	for _, forbidden := range []string{
		"兼容 `v1.0.0`",
		"发布新的 `v1.0.0`",
		"CreatorService",
		"RestaurantService",
		"open_id",
	} {
		if strings.Contains(readme, forbidden) {
			t.Fatalf("README.md contains forbidden %q:\n%s", forbidden, readme)
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
