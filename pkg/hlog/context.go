package hlog

import "context"

type contextKey struct{}

// NewContext 返回携带 Logger 的 Context；logger 为 nil 时使用默认 Logger。
func NewContext(ctx context.Context, logger *Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if logger == nil {
		logger = L()
	}
	return context.WithValue(ctx, contextKey{}, logger)
}

// FromContext 返回 Context 中的 Logger，未注入时使用默认 Logger。
func FromContext(ctx context.Context) *Logger {
	if ctx != nil {
		if logger, ok := ctx.Value(contextKey{}).(*Logger); ok && logger != nil {
			return logger
		}
	}
	return L()
}
