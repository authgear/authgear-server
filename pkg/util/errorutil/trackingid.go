package errorutil

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/trace"
)

func FormatTrackingID(ctx context.Context) string {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return ""
	}
	return fmt.Sprintf("%s-%s", sc.TraceID().String(), sc.SpanID().String())
}
