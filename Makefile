GOMOD=vendor

cli:
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/features cmd/features/main.go
