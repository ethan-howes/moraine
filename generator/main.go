package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// collectorEndpoint reads OTEL_EXPORTER_OTLP_ENDPOINT or returns the default.
func collectorEndpoint() string {
	if ep := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); ep != "" {
		return ep
	}
	return "localhost:4317"
}

// setupTracer initializes a TracerProvider for a named service and returns
// a shutdown function. Each simulated Talus service gets its own provider
// so traces appear as separate services in the collector output.
func setupTracer(ctx context.Context, conn *grpc.ClientConn, serviceName string) (*sdktrace.TracerProvider, error) {
	exp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("create exporter for %s: %w", serviceName, err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("0.1.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create resource for %s: %w", serviceName, err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	return tp, nil
}

func main() {
	ctx := context.Background()

	conn, err := grpc.NewClient(
		collectorEndpoint(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("dial collector: %v", err)
	}
	defer conn.Close()

	// one TracerProvider per simulated Talus service
	services := []string{
		"talus-ingestion",
		"talus-terrain",
		"talus-hazard",
		"talus-gateway",
	}

	providers := make(map[string]*sdktrace.TracerProvider)
	for _, svc := range services {
		tp, err := setupTracer(ctx, conn, svc)
		if err != nil {
			log.Fatalf("setup tracer: %v", err)
		}
		providers[svc] = tp
		defer tp.Shutdown(ctx)
	}

	log.Printf("generator started, emitting traces to %s", collectorEndpoint())

	// emit a realistic request cycle every 3 seconds
	for {
		emitRequestCycle(ctx, providers)
		time.Sleep(3 * time.Second)
	}
}

// emitRequestCycle simulates a full Talus request: gateway receives an
// analyze request, calls hazard, which calls terrain, which runs CUDA.
func emitRequestCycle(ctx context.Context, providers map[string]*sdktrace.TracerProvider) {

	// gateway handles the incoming request
	gatewayTracer := providers["talus-gateway"].Tracer("talus-gateway")
	gCtx, gatewaySpan := gatewayTracer.Start(ctx, "POST /analyze")
	gatewaySpan.SetAttributes(
		attribute.String("http.method", "POST"),
		attribute.String("http.route", "/analyze"),
		attribute.Int("http.status_code", 200),
	)

	// gateway calls hazard
	_, hazardSpan := providers["talus-hazard"].Tracer("talus-hazard").Start(gCtx, "POST /analyze")
	hazardSpan.SetAttributes(
		attribute.String("http.method", "POST"),
		attribute.String("http.route", "/analyze"),
		attribute.Int("route_id", 1),
	)
	simulateWork(50, 120) // hazard analysis takes 50-120ms

	// hazard calls terrain for GPU processing
	_, terrainSpan := providers["talus-terrain"].Tracer("talus-terrain").Start(gCtx, "POST /tile")
	terrainSpan.SetAttributes(
		attribute.String("http.method", "POST"),
		attribute.String("http.route", "/tile"),
	)

	// terrain invokes CUDA binary
	_, cudaSpan := providers["talus-terrain"].Tracer("talus-terrain").Start(gCtx, "cuda.terrain.run")
	gpuMs := randFloat(300, 600)
	cpuMs := gpuMs * randFloat(8, 11) // realistic speedup ratio
	cudaSpan.SetAttributes(
		attribute.Int64("cuda.cells", 116_900_000),
		attribute.Float64("cuda.slope_aspect.gpu_ms", gpuMs),
		attribute.Float64("cuda.slope_aspect.cpu_ms", cpuMs),
		attribute.Float64("cuda.tri.gpu_ms", gpuMs*0.55),
		attribute.Float64("cuda.tri.cpu_ms", cpuMs*0.27),
	)
	simulateWork(int(gpuMs), int(gpuMs)+50)
	cudaSpan.End()

	terrainSpan.End()
	hazardSpan.End()

	// occasionally simulate an error
	if rand.Float64() < 0.05 {
		gatewaySpan.SetStatus(codes.Error, "upstream timeout")
		gatewaySpan.SetAttributes(attribute.Int("http.status_code", 504))
	}

	gatewaySpan.End()
}

func simulateWork(minMs, maxMs int) {
	d := minMs + rand.Intn(maxMs-minMs)
	time.Sleep(time.Duration(d) * time.Millisecond)
}

func randFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}
