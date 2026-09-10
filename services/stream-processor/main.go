package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/ethan-howes/moraine/proto"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "localhost:9092"
	}

	aggregatorAddr := os.Getenv("AGGREGATOR_ADDR")
	if aggregatorAddr == "" {
		aggregatorAddr = "localhost:50051"
	}

	alertAddr := os.Getenv("ALERT_ENGINE_ADDR")
	if alertAddr == "" {
		alertAddr = "localhost:50052"
	}

	// connect to aggregator
	aggConn, err := grpc.NewClient(aggregatorAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Error("failed to connect to aggregator", "error", err)
		os.Exit(1)
	}
	defer aggConn.Close()
	aggClient := pb.NewAggregatorServiceClient(aggConn)

	// connect to alert engine
	alertConn, err := grpc.NewClient(alertAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Error("failed to connect to alert engine", "error", err)
		os.Exit(1)
	}
	defer alertConn.Close()
	alertClient := pb.NewAlertEngineServiceClient(alertConn)

	// kafka reader consuming from talus.traces
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaBroker},
		Topic:    "talus.traces",
		GroupID:  "moraine-stream-processor",
		MinBytes: 1,
		MaxBytes: 10e6,
		MaxWait:  1 * time.Second,
	})
	defer reader.Close()

	logger.Info("stream processor started",
		"kafka_broker", kafkaBroker,
		"aggregator", aggregatorAddr,
		"alert_engine", alertAddr,
	)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				logger.Info("shutting down")
				return
			}
			logger.Error("failed to read message", "error", err)
			continue
		}

		event, err := parseEvent(msg.Value)
		if err != nil {
			logger.Error("failed to parse event, dropping", "error", err)
			continue
		}

		// call aggregator — drop on failure
		aggCtx, aggCancel := context.WithTimeout(ctx, 3*time.Second)
		_, err = aggClient.RecordMetric(aggCtx, event)
		aggCancel()
		if err != nil {
			logger.Error("aggregator call failed, dropping", "error", err)
		}

		// call alert engine — drop on failure
		alertCtx, alertCancel := context.WithTimeout(ctx, 3*time.Second)
		resp, err := alertClient.EvaluateEvent(alertCtx, event)
		alertCancel()
		if err != nil {
			logger.Error("alert engine call failed, dropping", "error", err)
		} else if resp.AlertFired {
			logger.Info("alert fired", "reason", resp.Reason)
		}
	}
}

// parseEvent converts a raw Kafka message body into a TelemetryEvent.
// For now this is a stub — in Phase 5 this will deserialize real OTLP data.
func parseEvent(data []byte) (*pb.TelemetryEvent, error) {
	return &pb.TelemetryEvent{
		ServiceName:  "talus-unknown",
		SpanName:     "unknown",
		DurationMs:   0,
		StatusCode:   200,
		Timestamp:    time.Now().UnixMilli(),
		TraceId:      "",
		ErrorMessage: "",
	}, nil
}
