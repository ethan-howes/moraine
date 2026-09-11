output "eks_cluster_name" {
  description = "EKS cluster name"
  value       = aws_eks_cluster.moraine.name
}

output "eks_cluster_endpoint" {
  description = "EKS cluster API endpoint"
  value       = aws_eks_cluster.moraine.endpoint
}

output "msk_bootstrap_brokers" {
  description = "MSK bootstrap broker connection string"
  value       = aws_msk_cluster.moraine.bootstrap_brokers_tls
}

output "ecr_stream_processor_url" {
  description = "ECR URL for the stream processor image"
  value       = aws_ecr_repository.stream_processor.repository_url
}

output "ecr_aggregator_url" {
  description = "ECR URL for the aggregator image"
  value       = aws_ecr_repository.aggregator.repository_url
}

output "ecr_alert_engine_url" {
  description = "ECR URL for the alert engine image"
  value       = aws_ecr_repository.alert_engine.repository_url
}
