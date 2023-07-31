package main

import (
	"database/sql"
	"log"

	"github.com/PYTNAG/simpletodo/api"
	db "github.com/PYTNAG/simpletodo/db/sqlc"

	_ "github.com/lib/pq"
)

const (
	dbDriver   = "postgres"
	dbSource   = "postgresql://root:mysecret@localhost:5432/simple_todo?sslmode=disable"
	serverAddr = "localhost:8080"
)

func main() {
	var err error
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("Cannot connect to db: ", err)
	}

	store := db.NewStore(conn)

	server := api.NewServer(store)

	err = server.Start(serverAddr)
	if err != nil {
		log.Fatal("can't start server: ", err)
	}
}
