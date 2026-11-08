.PHONY: run build test lint frontend backend

frontend:
	cd web && npm install && npm run build

backend:
	go build -o bin/ultima-pista ./cmd/server

build: frontend backend

run: build
	./bin/ultima-pista -addr :8080 -db data/ultima-pista.db -static web/dist

test:
	go test ./... -count=1

lint:
	cd web && npm run lint
