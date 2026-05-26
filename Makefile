.PHONY: dev build run clean

dev:
	cd web && npm run dev &
	go run .

build:
	cd web && npm run build
	go build -o bin/cloudalbum .

run: build
	./bin/cloudalbum

clean:
	rm -rf bin/ web/dist/ data/
