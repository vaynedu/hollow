package generator

import "testing"

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
