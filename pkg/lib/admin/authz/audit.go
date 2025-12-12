package authz

import "context"

type auditContextKeyType struct{}

var auditContextKey = auditContextKeyType{}

const (
	JWTKeyAuditContext string = "audit_context"
)

type T struct {
	AuditContext map[string]any
}

func WithAdminAuthzAudit(ctx context.Context, content map[string]any) context.Context {
	v, ok := ctx.Value(auditContextKey).(*T)
	if ok {
		v.AuditContext = content
		return ctx
	}
	t := &T{AuditContext: content}
	return context.WithValue(ctx, auditContextKey, t)
}

func GetAdminAuthzAudit(ctx context.Context) map[string]any {
	v, ok := ctx.Value(auditContextKey).(*T)
	if ok {
		return v.AuditContext
	}
	return nil
}
