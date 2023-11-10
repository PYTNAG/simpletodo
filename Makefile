postgres:
	docker run --name todo_db --network todo-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=mysecret -d postgres:15.3

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

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/PYTNAG/simpletodo/db/sqlc Store

proto:
	rm -f pb/*.go
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
    proto/*.proto

evans:
	evans --host 0.0.0.0 --port 8090 -r repl

.PHONY: postgres start stop createdb dropdb migrateup migratedown docker_sqlc test server mock proto evans
