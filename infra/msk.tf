# Security group for MSK — controls what can connect to the Kafka brokers.
# Only allows traffic from within the VPC so brokers are not publicly reachable.
resource "aws_security_group" "msk" {
  name        = "${var.project}-msk-sg"
  description = "Security group for Moraine MSK cluster"
  vpc_id      = aws_vpc.moraine.id

  ingress {
    description = "Kafka plaintext from within VPC"
    from_port   = 9092
    to_port     = 9092
    protocol    = "tcp"
    cidr_blocks = [aws_vpc.moraine.cidr_block]
  }

  ingress {
    description = "Kafka TLS from within VPC"
    from_port   = 9094
    to_port     = 9094
    protocol    = "tcp"
    cidr_blocks = [aws_vpc.moraine.cidr_block]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name    = "${var.project}-msk-sg"
    Project = var.project
  }
}

# MSK configuration — sets the Kafka broker properties.
resource "aws_msk_configuration" "moraine" {
  name              = "${var.project}-msk-config"
  kafka_versions    = ["3.6.0"]

  server_properties = <<-EOF
    auto.create.topics.enable=false
    default.replication.factor=1
    min.insync.replicas=1
    num.partitions=1
    log.retention.hours=24
  EOF
}

# MSK cluster — the managed Kafka service.
# Single broker in one AZ for dev. Production would use 3 brokers across 3 AZs.
resource "aws_msk_cluster" "moraine" {
  cluster_name           = "${var.project}-kafka"
  kafka_version          = "3.6.0"
  number_of_broker_nodes = 1

  broker_node_group_info {
    instance_type  = "kafka.t3.small"
    client_subnets = [aws_subnet.private[0].id]

    storage_info {
      ebs_storage_info {
        volume_size = 20
      }
    }

    security_groups = [aws_security_group.msk.id]
  }

  configuration_info {
    arn      = aws_msk_configuration.moraine.arn
    revision = aws_msk_configuration.moraine.latest_revision
  }

  encryption_info {
    encryption_in_transit {
      client_broker = "PLAINTEXT"
      in_cluster    = false
    }
  }

  tags = { Project = var.project }
}
