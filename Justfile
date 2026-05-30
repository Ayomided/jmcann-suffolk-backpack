set dotenv-load
db_path := env("BACKPACK_APP_DB", "test.db")

@gen-jwt:
    openssl rand -base64 32

@test:
    go test -v ./...

@build:
    go build -o backpack-app ./cmd/backend/main.go

[no-exit-message]
@dev:
    go run ./cmd/backend/main.go -addr :3000 -db_path={{ db_path }}

@seed: build
    go run ./cmd/backend/main.go -db_path={{ db_path }} -seed

@migrate: build
    go run ./cmd/backend/main.go -db_path={{ db_path }} -migrate

@clean:
    rm {{ db_path }}
    rm backpack-app
