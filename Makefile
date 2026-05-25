.PHONY: dev build run clean

dev:
	cd web && npm run dev &
	go run cmd/server/main.go

build:
	cd web && npm run build
	go build -o bin/cloudalbum cmd/server/main.go

run: build
	./bin/cloudalbum

clean:
	rm -rf bin/ web/dist/ data/
