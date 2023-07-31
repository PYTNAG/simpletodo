postgres:
	docker run --name todo_db -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=mysecret -d postgres:15.3

start:
	docker start todo_db

stop:
	docker stop todo_db

createdb:
	docker exec -it todo_db createdb --username=root --owner=root simple_todo

dropdb:
	docker exec -it todo_db dropdb simple_todo

migrateup:
	migrate -path db/migration -database "postgresql://root:mysecret@localhost:5432/simple_todo?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:mysecret@localhost:5432/simple_todo?sslmode=disable" -verbose down

docker_sqlc:
	docker run --rm -v .:/src -w /src kjconroy/sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

.PHONY: postgres start stop createdb dropdb migrateup migratedown docker_sqlc test server
