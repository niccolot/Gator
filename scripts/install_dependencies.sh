#! usr/bin/bash

go get github.com/google/uuid
go get github.com/joho/godotenv
go get github.com/lib/pq
go get github.com/mattn/go-runewidth
go get github.com/peterh/liner
go get golang.org/x/crypto
go get golang.org/x/sys
go install go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

sudo apt update
sudo apt install postgresql postgresql-contrib
sudo passwd postgres