PROJECT_NAME=marketflow

EXCH1=images/exchange1_amd64.tar
EXCH2=images/exchange2_amd64.tar
EXCH3=images/exchange3_amd64.tar

DC=docker-compose 

load:
	@echo "Loading exchange images..."
	docker load -i $(EXCH1)
	docker load -i $(EXCH2)
	docker load -i $(EXCH3)

build:
	@echo "Building the project..."
	go build -o marketflow ./cmd/marketflow/main.go

up:
	@echo "Starting $(PROJECT_NAME)..."
	$(DC) up --build

down:
	@echo "Stopping $(PROJECT_NAME)..."
	$(DC) down

restart: down up

nuke:
	@echo "Removing all containers, networks, and volumes..."
	$(DC) down -v

