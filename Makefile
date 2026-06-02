.PHONY: api worker up down migrate-up migrate-down

api:
	go run cmd/api/main.go

worker:
	go run cmd/worker/main.go

up:
	docker-compose up -d

down:
	docker-compose down

migrate-up:
	migrate -path migrations -database "postgresql://root:secretpassword@localhost:5432/worker_queue?sslmode=disable" -verbose up

migrate-down:
	migrate -path migrations -database "postgresql://root:secretpassword@localhost:5432/worker_queue?sslmode=disable" -verbose down
