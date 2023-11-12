package main

import (
	"database/sql"
	"log"
	"net"

	db "github.com/PYTNAG/simpletodo/db/sqlc"
	"github.com/PYTNAG/simpletodo/gapi"
	"github.com/PYTNAG/simpletodo/pb"
	"github.com/PYTNAG/simpletodo/util"
	"github.com/golang-migrate/migrate/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

	runDBMigration(cfg.MigrationURL, cfg.DBSource)

	store := db.NewStore(conn)

	startGrpcServer(cfg, store)
}

func runDBMigration(migrationURL, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal("cannot create a new migrate instance: ", err)
	}

	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("failed to run migrate up: ", err)
	}

	log.Print("db migrated successfully")
}

func startGrpcServer(cfg util.Config, store db.Store) {
	grpcServer := grpc.NewServer()

	server, err := gapi.NewServer(cfg, store)
	if err != nil {
		log.Fatal("cannot create server: ", err)
	}

	pb.RegisterSimpleTODOServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", cfg.GrpcServerAddr)
	if err != nil {
		log.Fatal("cannot create listener: ", err)
	}

	log.Printf("start gRPC server at %s", listener.Addr())

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start gRPC server: ", err)
	}
}
