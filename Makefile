TALUS_DIR := /home/yart/Desktop/personal-projects/talus
SERVICES := postgres ingestion hazard gateway otel-collector otel-ingestion otel-hazard otel-gateway generator kafka aggregator alert-engine stream-processor

up:
	docker compose -f $(TALUS_DIR)/deployments/docker-compose.yml -f docker-compose.moraine.yml up -d $(SERVICES)

down:
	docker compose -f $(TALUS_DIR)/deployments/docker-compose.yml -f docker-compose.moraine.yml down

logs-collector:
	docker logs deployments-otel-collector-1 --follow

logs-generator:
	docker logs deployments-generator-1 --follow

infra-up:
	cd infra && terraform apply -auto-approve

infra-down:
	cd infra && terraform destroy -auto-approve
