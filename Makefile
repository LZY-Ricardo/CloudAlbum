.PHONY: dev dev-backend dev-frontend build run docker docker-down clean

dev:
	@make dev-frontend &
	@make dev-backend

dev-backend:
	go run .

dev-frontend:
	cd web && npm run dev

build:
	cd web && npm ci && npm run build
	CGO_ENABLED=1 go build -o bin/cloudalbum .

run: build
	./bin/cloudalbum

docker:
	docker compose up -d --build

docker-down:
	docker compose down

clean:
	rm -rf bin/ cloudalbum web/dist/ data/
