TALUS_DIR := /home/yart/Desktop/personal-projects/talus
SERVICES := postgres ingestion hazard gateway otel-collector otel-ingestion otel-hazard otel-gateway generator kafka aggregator alert-engine stream-processor

up:
	docker compose -f $(TALUS_DIR)/deployments/docker-compose.yml -f docker-compose.moraine.yml up -d $(SERVICES)
	sleep 15
	docker exec moraine-kafka /opt/kafka/bin/kafka-topics.sh --create --if-not-exists --topic talus.traces --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1
	docker exec moraine-kafka /opt/kafka/bin/kafka-topics.sh --create --if-not-exists --topic talus.metrics --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1
	docker exec moraine-kafka /opt/kafka/bin/kafka-topics.sh --create --if-not-exists --topic talus.logs --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1
down:
	docker compose -f $(TALUS_DIR)/deployments/docker-compose.yml -f docker-compose.moraine.yml down

logs-collector:
	docker logs deployments-otel-collector-1 --follow

logs-generator:
	docker logs deployments-generator-1 --follow

logs-sp:
	docker logs deployments-stream-processor-1 --follow

infra-up:
	cd infra && terraform apply -auto-approve

infra-down:
	cd infra && terraform destroy -auto-approve
