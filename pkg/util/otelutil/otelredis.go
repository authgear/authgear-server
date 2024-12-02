package otelutil

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/semconv/v1.27.0"
)

const otelredisMetricName = "github.com/redis/go-redis/extra/redisotel"

// InstrumentMetrics is based on https://github.com/redis/go-redis/blob/v9.7.0/extra/redisotel/metrics.go
// With the following modifications
//
// 1. Added the ability to get extra attributes from context.
// 2. Convert some metric to use the semconv v1.28.0
func OtelRedisInstrumentMetrics(client *redis.Client) error {
	metricProvider := otel.GetMeterProvider()

	meter := metricProvider.Meter(
		otelredisMetricName,
		metric.WithInstrumentationVersion("semver:"+redis.Version()),
	)

	options := client.Options()

	baseAttrs := []attribute.KeyValue{
		semconv.DBSystemKey.String("redis"),
		semconv.DBNamespaceKey.String(strconv.Itoa(options.DB)),
	}

	if host, portStr, err := net.SplitHostPort(options.Addr); err == nil {
		if port, err := strconv.Atoi(portStr); err == nil {
			baseAttrs = append(baseAttrs, semconv.ServerAddressKey.String(host))
			baseAttrs = append(baseAttrs, semconv.ServerPortKey.Int(port))
		}
	}

	idleMax, err := meter.Int64ObservableUpDownCounter(
		semconv.DBClientConnectionIdleMaxName,
		metric.WithDescription(semconv.DBClientConnectionIdleMaxDescription),
	)
	if err != nil {
		return err
	}

	idleMin, err := meter.Int64ObservableUpDownCounter(
		semconv.DBClientConnectionIdleMinName,
		metric.WithDescription(semconv.DBClientConnectionIdleMinDescription),
	)
	if err != nil {
		return err
	}

	connsMax, err := meter.Int64ObservableUpDownCounter(
		semconv.DBClientConnectionMaxName,
		metric.WithDescription(semconv.DBClientConnectionMaxDescription),
	)
	if err != nil {
		return err
	}

	usage, err := meter.Int64ObservableUpDownCounter(
		semconv.DBClientConnectionCountName,
		metric.WithDescription(semconv.DBClientConnectionCountDescription),
	)
	if err != nil {
		return err
	}

	timeouts, err := meter.Int64ObservableCounter(
		semconv.DBClientConnectionsTimeoutsName,
		metric.WithDescription(semconv.DBClientConnectionsTimeoutsDescription),
	)
	if err != nil {
		return err
	}

	_, err = meter.RegisterCallback(
		func(ctx context.Context, o metric.Observer) error {
			options := client.Options()
			stats := client.PoolStats()

			{
				labels := append([]attribute.KeyValue(nil), baseAttrs...)
				o.ObserveInt64(idleMax, int64(options.MaxIdleConns), metric.WithAttributes(labels...))
			}
			{
				labels := append([]attribute.KeyValue(nil), baseAttrs...)
				o.ObserveInt64(idleMin, int64(options.MinIdleConns), metric.WithAttributes(labels...))
			}
			{
				labels := append([]attribute.KeyValue(nil), baseAttrs...)
				o.ObserveInt64(connsMax, int64(options.PoolSize), metric.WithAttributes(labels...))
			}
			{
				labels := append([]attribute.KeyValue(nil), baseAttrs...)
				labels = append(labels, attribute.String("state", "idle"))
				o.ObserveInt64(usage, int64(stats.IdleConns), metric.WithAttributes(labels...))
			}
			{
				labels := append([]attribute.KeyValue(nil), baseAttrs...)
				labels = append(labels, attribute.String("state", "used"))
				o.ObserveInt64(usage, int64(stats.TotalConns-stats.IdleConns), metric.WithAttributes(labels...))
			}
			{
				labels := append([]attribute.KeyValue(nil), baseAttrs...)
				o.ObserveInt64(timeouts, int64(stats.Timeouts), metric.WithAttributes(labels...))
			}
			return nil
		},
		idleMax,
		idleMin,
		connsMax,
		usage,
		timeouts,
	)
	if err != nil {
		return err
	}

	createTime, err := meter.Float64Histogram(
		semconv.DBClientConnectionCreateTimeName,
		metric.WithDescription(semconv.DBClientConnectionCreateTimeDescription),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	useTime, err := meter.Float64Histogram(
		semconv.DBClientConnectionUseTimeName,
		metric.WithDescription(semconv.DBClientConnectionUseTimeDescription),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	client.AddHook(&otelRedisMetricsHook{
		createTime: createTime,
		useTime:    useTime,
		baseAttrs:  baseAttrs,
	})

	return nil
}

type otelRedisMetricsHook struct {
	createTime metric.Float64Histogram
	useTime    metric.Float64Histogram
	baseAttrs  []attribute.KeyValue
}

var _ redis.Hook = (*otelRedisMetricsHook)(nil)

func (h *otelRedisMetricsHook) DialHook(hook redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		start := time.Now()
		conn, err := hook(ctx, network, addr)
		dur := time.Since(start)

		attrs := append([]attribute.KeyValue(nil), h.baseAttrs...)
		attrs = append(attrs, statusAttr(err))

		h.createTime.Record(ctx, seconds(dur), metric.WithAttributes(attrs...))
		return conn, err
	}
}

func (h *otelRedisMetricsHook) ProcessHook(hook redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		err := hook(ctx, cmd)
		dur := time.Since(start)

		attrs := append([]attribute.KeyValue(nil), h.baseAttrs...)
		attrs = append(attrs, attribute.String("type", "command"))
		attrs = append(attrs, semconv.DBOperationName(cmd.FullName()))
		attrs = append(attrs, statusAttr(err))

		h.useTime.Record(ctx, seconds(dur), metric.WithAttributes(attrs...))
		return err
	}
}

func (h *otelRedisMetricsHook) ProcessPipelineHook(
	hook redis.ProcessPipelineHook,
) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		start := time.Now()
		err := hook(ctx, cmds)
		dur := time.Since(start)

		attrs := append([]attribute.KeyValue(nil), h.baseAttrs...)
		attrs = append(attrs, attribute.String("type", "pipeline"))
		attrs = append(attrs, statusAttr(err))

		h.useTime.Record(ctx, seconds(dur), metric.WithAttributes(attrs...))
		return err
	}
}

func seconds(d time.Duration) float64 {
	return float64(d) / float64(time.Second)
}

func statusAttr(err error) attribute.KeyValue {
	if err != nil {
		return attribute.String("status", "error")
	}
	return attribute.String("status", "ok")
}
