package main

import (
	"context"
	"log/slog"
	"net"
	"os"

	"google.golang.org/grpc"

	pb "github.com/ethan-howes/moraine/proto"
)

type aggregatorServer struct {
	pb.UnimplementedAggregatorServiceServer
	logger *slog.Logger
}

func (s *aggregatorServer) RecordMetric(ctx context.Context, event *pb.TelemetryEvent) (*pb.RecordMetricResponse, error) {
	s.logger.Info("received metric",
		"service", event.ServiceName,
		"span", event.SpanName,
		"duration_ms", event.DurationMs,
		"status_code", event.StatusCode,
		"error", event.ErrorMessage,
	)
	return &pb.RecordMetricResponse{Accepted: true}, nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	port := os.Getenv("AGGREGATOR_PORT")
	if port == "" {
		port = "50051"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	srv := grpc.NewServer()
	pb.RegisterAggregatorServiceServer(srv, &aggregatorServer{logger: logger})

	logger.Info("aggregator listening", "port", port)
	if err := srv.Serve(lis); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
