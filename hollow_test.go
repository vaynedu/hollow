package hollow

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewAppUsesEnvironmentPrefix(t *testing.T) {
	dir := t.TempDir()
	content := "server:\n  host: 127.0.0.1:8080\n  shutdown_timeout: 10s\nlog:\n  level: info\n  output_mode: console\n"
	if err := os.WriteFile(filepath.Join(dir, "conf.yaml"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LIFELOG_SERVER_HOST", "127.0.0.1:9090")

	app, err := NewApp(AppOption{ConfigPath: dir, ConfigName: "conf", EnvPrefix: "LIFELOG"})
	if err != nil {
		t.Fatal(err)
	}
	if got := app.config.Server.Host; got != "127.0.0.1:9090" {
		t.Fatalf("server.host=%q，期望环境变量值", got)
	}
}
