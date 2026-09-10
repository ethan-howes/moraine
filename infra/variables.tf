variable "aws_region" {
  description = "AWS region for Moraine infrastructure deployment"
  type        = string
  default     = "us-east-2"
}

variable "project" {
  description = "Project name to prefix all resources"
  type        = string
  default     = "moraine"
}

variable "kafka_topics" {
  description = "Kafka topics for Talus telemetry"
  type        = list(string)
  default     = ["talus.traces", "talus.metrics", "talus.logs"]
}
