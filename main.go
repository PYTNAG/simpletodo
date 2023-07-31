package main

import (
	"database/sql"
	"log"

	"github.com/PYTNAG/simpletodo/api"
	db "github.com/PYTNAG/simpletodo/db/sqlc"
	"github.com/PYTNAG/simpletodo/util"

	_ "github.com/lib/pq"
)

func main() {
	cfg, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("Cannot load config: ", err)
	}

	conn, err := sql.Open(cfg.DBDriver, cfg.DBSource)
	if err != nil {
		log.Fatal("Cannot connect to db: ", err)
	}

	store := db.NewStore(conn)

	server := api.NewServer(store)

	err = server.Start(cfg.ServerAddr)
	if err != nil {
		log.Fatal("Cannot start server: ", err)
	}
}
