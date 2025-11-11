# Build and generate the go binary
build:
	go mod tidy
	go build
	go install

# Open log file
logs:
	tail -f 50 ~/.coconut/logs/coconut.log

# Clear the database
clear_db:
	rm ~/.coconut/coconut.db

# Dump the database contents
db_dump:
	go run scripts/db_dump.go

.PHONY: build logs clear_db db_dump