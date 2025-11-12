build:
	go mod tidy
	go install

test:
	go test ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

logs:
	tail -f ~/.coconut/logs/coconut.log

clear_db:
	@read -p "Delete all passwords? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	rm -f ~/.coconut/coconut.db

.PHONY: build test coverage logs clear_db