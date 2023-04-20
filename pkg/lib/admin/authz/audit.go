package authz

import "context"

type auditContextKeyType struct{}

var auditContextKey = auditContextKeyType{}

const (
	JWTKeyAuditContext string = "audit_context"
)

type T struct {
	AuditContext interface{}
}

func WithAdminAuthzAudit(ctx context.Context, content interface{}) context.Context {
	v, ok := ctx.Value(auditContextKey).(*T)
	if ok {
		v.AuditContext = content
		return ctx
	}
	t := &T{AuditContext: content}
	return context.WithValue(ctx, auditContextKey, t)
}

func GetAdminAuthzAudit(ctx context.Context) interface{} {
	v, ok := ctx.Value(auditContextKey).(*T)
	if ok {
		return v.AuditContext
	}
	return nil
}
