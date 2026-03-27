package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vaynedu/hollow/internal/config"
	"go.uber.org/zap"
)

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{
			name: "default config",
			cfg:  nil,
			wantErr: false,
		},
		{
			name: "with console config",
			cfg: &config.Config{
				Log: config.LogConfig{
					LogLevel:   "info",
					OutputMode: "console",
				},
			},
			wantErr: false,
		},
		{
			name: "with file config",
			cfg: &config.Config{
				Log: config.LogConfig{
					LogLevel:    "debug",
					OutputMode:  "file",
					LogFileName: "test.log",
					MaxSize:     10,
					MaxAge:      7,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, err := InitLogger(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, log)
			}
		})
	}
}

func TestGetLogger(t *testing.T) {
	// Reset logger to nil for testing
	logger = nil
	
	log := GetLogger()
	assert.NotNil(t, log)
}

func TestLogFunctions(t *testing.T) {
	// Initialize logger
	InitLogger(nil)

	// Test formatted logging functions
	Debugf("test debug: %s", "message")
	Infof("test info: %s", "message")
	Warnf("test warn: %s", "message")
	Errorf("test error: %s", "message")

	// Test logging with fields
	Debug("debug message", zap.String("key", "value"))
	Info("info message", zap.String("key", "value"))
	Warn("warn message", zap.String("key", "value"))
	Error("error message", zap.String("key", "value"))
}

func TestWithFields(t *testing.T) {
	InitLogger(nil)
	
	log := WithFields(zap.String("request_id", "12345"))
	assert.NotNil(t, log)
	log.Info("test with fields")
}
