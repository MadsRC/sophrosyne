// Sophrosyne
//   Copyright (C) 2024  Mads R. Havmand
//
// This program is free software: you can redistribute it and/or modify
//   it under the terms of the GNU Affero General Public License as published by
//   the Free Software Foundation, either version 3 of the License, or
//   (at your option) any later version.
//
//   This program is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//   GNU Affero General Public License for more details.
//
//   You should have received a copy of the GNU Affero General Public License
//   along with this program.  If not, see <http://www.gnu.org/licenses/>.

package otel

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/madsrc/sophrosyne"
)

// SetupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTelSDK(ctx context.Context, config *sophrosyne.Config) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	res, err := resource.New(ctx, resource.WithAttributes(
		semconv.ServiceNameKey.String("sophrosyne"),
		semconv.ServiceVersionKey.String("0.0.0"),
	),
	)
	if err != nil {
		return nil, err
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	if config.Tracing.Enabled {
		// Set up trace provider.
		tracerProvider, err := newTraceProvider(ctx, config, res)
		if err != nil {
			handleErr(err)
			return shutdown, err
		}
		shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
		otel.SetTracerProvider(tracerProvider)
	}

	if config.Metrics.Enabled {
		// Set up meter provider.
		meterProvider, err := newMeterProvider(ctx, config, res)
		if err != nil {
			handleErr(err)
			return shutdown, err
		}
		shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
		otel.SetMeterProvider(meterProvider)
	}

	return
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(ctx context.Context, config *sophrosyne.Config, res *resource.Resource) (*sdkTrace.TracerProvider, error) {
	var traceExporter sdkTrace.SpanExporter
	var err error
	if config.Tracing.Output == sophrosyne.OtelOutputHTTP {
		traceExporter, err = otlptracehttp.New(ctx)
	} else {
		traceExporter, err = stdouttrace.New()
	}
	if err != nil {
		return nil, err
	}

	traceProvider := sdkTrace.NewTracerProvider(
		sdkTrace.WithBatcher(traceExporter,
			sdkTrace.WithBatchTimeout(time.Duration(config.Tracing.Batch.Timeout)*time.Second)),
		sdkTrace.WithResource(res),
	)
	return traceProvider, nil
}

func newMeterProvider(ctx context.Context, config *sophrosyne.Config, res *resource.Resource) (*sdkMetric.MeterProvider, error) {
	var metricExporter sdkMetric.Exporter
	var err error
	if config.Metrics.Output == sophrosyne.OtelOutputHTTP {
		metricExporter, err = otlpmetrichttp.New(ctx)
	} else {
		metricExporter, err = stdoutmetric.New()
	}
	if err != nil {
		return nil, err
	}

	meterProvider := sdkMetric.NewMeterProvider(
		sdkMetric.WithReader(sdkMetric.NewPeriodicReader(metricExporter,
			sdkMetric.WithInterval(time.Duration(config.Metrics.Interval)*time.Second))),
		sdkMetric.WithResource(res),
	)
	return meterProvider, nil
}

type Span struct {
	span trace.Span
}

func (s *Span) End() {
	s.span.End()
}

type OtelService struct {
	panicMeter metric.Meter
	panicCnt   metric.Int64Counter
}

func NewOtelService() (*OtelService, error) {
	panicMeter := otel.Meter("panics")
	panicCnt, err := panicMeter.Int64Counter("panics",
		metric.WithDescription("Number of panics"),
		metric.WithUnit("{{total}}"))
	if err != nil {
		return nil, err
	}
	return &OtelService{panicMeter: panicMeter, panicCnt: panicCnt}, nil
}

func (o *OtelService) RecordPanic(ctx context.Context) {
	o.panicCnt.Add(ctx, 1)
}

func (o *OtelService) StartSpan(ctx context.Context, name string) (context.Context, sophrosyne.Span) {
	ctx, span := otel.Tracer("internal/otel").Start(ctx, name)
	return ctx, &Span{span: span}
}

func (o *OtelService) GetTraceID(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		traceID := spanCtx.TraceID()
		return traceID.String()
	}
	return ""
}

func (o *OtelService) NewHTTPHandler(operation string, handler http.Handler) http.Handler {
	return otelhttp.NewHandler(handler, operation)
}

func (o *OtelService) WithRouteTag(pattern string, handler http.Handler) http.Handler {
	return otelhttp.WithRouteTag(pattern, handler)
}
