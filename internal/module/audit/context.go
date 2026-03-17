package audit

import (
	"context"
)

type AuditContext struct {
	UserID    uint
	RequestID string
	Method    string
	Path      string
	IPAddress string
	UserAgent string
}

type contextKey struct{}

func FromContext(ctx context.Context) *AuditContext {
	if val, ok := ctx.Value(contextKey{}).(*AuditContext); ok {
		return val
	}
	return nil
}

func NewContext(ctx context.Context, auditCtx *AuditContext) context.Context {
	return context.WithValue(ctx, contextKey{}, auditCtx)
}
