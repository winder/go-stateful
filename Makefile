test:
	go test ./...

lint:
	test -z "$$(gofmt -l .)"
	go vet ./...

coverage:
	go test -coverprofile=cover.out ./...
	go tool cover -func=cover.out
