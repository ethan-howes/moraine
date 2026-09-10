package main

import (
	"context"
	"bytes"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"os"

	"google.golang.org/grpc"

	pb "github.com/ethan-howes/moraine/proto"
)

type alertServer struct {
	pb.UnimplementedAlertEngineServiceServer
	logger     *slog.Logger
	webhookURL string
	threshold  float64
}

func (s *alertServer) EvaluateEvent(ctx context.Context, event *pb.TelemetryEvent) (*pb.EvaluateEventResponse, error) {
	if event.ErrorMessage != "" {
		s.logger.Info("error event received, firing alert",
			"service", event.ServiceName,
			"span", event.SpanName,
			"error", event.ErrorMessage,
		)
		err := s.fireWebhook(event)
		if err != nil {
			s.logger.Error("webhook failed", "error", err)
		}
		return &pb.EvaluateEventResponse{
			AlertFired: true,
			Reason:     "error detected: " + event.ErrorMessage,
		}, nil
	}

	if event.DurationMs > s.threshold {
		s.logger.Info("latency threshold breached, firing alert",
			"service", event.ServiceName,
			"span", event.SpanName,
			"duration_ms", event.DurationMs,
			"threshold_ms", s.threshold,
		)
		err := s.fireWebhook(event)
		if err != nil {
			s.logger.Error("webhook failed", "error", err)
		}
		return &pb.EvaluateEventResponse{
			AlertFired: true,
			Reason:     "latency threshold breached",
		}, nil
	}

	return &pb.EvaluateEventResponse{AlertFired: false}, nil
}

func (s *alertServer) fireWebhook(event *pb.TelemetryEvent) error {
	if s.webhookURL == "" {
		return nil
	}

	payload := map[string]any{
		"service":       event.ServiceName,
		"span":          event.SpanName,
		"duration_ms":   event.DurationMs,
		"error_message": event.ErrorMessage,
		"trace_id":      event.TraceId,
		"timestamp":     event.Timestamp,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(s.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	port := os.Getenv("ALERT_ENGINE_PORT")
	if port == "" {
		port = "50052"
	}

	webhookURL := os.Getenv("ALERT_WEBHOOK_URL")

	threshold := 1000.0

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	srv := grpc.NewServer()
	pb.RegisterAlertEngineServiceServer(srv, &alertServer{
		logger:     logger,
		webhookURL: webhookURL,
		threshold:  threshold,
	})

	logger.Info("alert engine listening", "port", port)
	if err := srv.Serve(lis); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
