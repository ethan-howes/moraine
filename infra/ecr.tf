# One ECR repository per Moraine service.
# ECR is AWS's managed Docker registry — EKS pulls images from here.

resource "aws_ecr_repository" "stream_processor" {
  name                 = "${var.project}/stream-processor"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = { Project = var.project }
}

resource "aws_ecr_repository" "aggregator" {
  name                 = "${var.project}/aggregator"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = { Project = var.project }
}

resource "aws_ecr_repository" "alert_engine" {
  name                 = "${var.project}/alert-engine"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = { Project = var.project }
}
