package hecode

import (
	"errors"
	"fmt"
	"testing"
)

func TestResponseResolution(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		code   int
		msg    string
	}{
		{"invalid param", fmt.Errorf("%w: bad id", ErrInvalidParam), 400, 1100, "invalid parameter"},
		{"unauthorized", ErrUnauthorized, 401, 1204, "unauthorized"},
		{"permission denied", ErrPermissionDenied, 403, 1202, "permission denied"},
		{"forbidden", ErrForbidden, 403, 1203, "forbidden access"},
		{"not found", ErrNotFound, 404, 1200, "resource not found"},
		{"plain error", errors.New("dsn=root:secret"), 500, 1001, "internal server error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp := Failure("req-1", tt.err)
			if status != tt.status || resp.Code != tt.code || resp.Msg != tt.msg {
				t.Fatalf("status=%d response=%+v", status, resp)
			}
			if resp.RequestID != "req-1" || resp.Data != nil {
				t.Fatalf("unexpected envelope: %+v", resp)
			}
		})
	}
}

func TestSuccessResponse(t *testing.T) {
	resp := Success("req-2", map[string]string{"status": "ok"})
	if resp.Code != 0 || resp.Msg != "success" || resp.RequestID != "req-2" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}
