package main

import (
	"database/sql"
	"net"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

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
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.DateTime,
	})

	cfg, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Msgf("Cannot load config: %s", err)
	}

	switch cfg.Env {
	case "development":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "production":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	conn, err := sql.Open(cfg.DBDriver, cfg.DBSource)
	if err != nil {
		log.Fatal().Msgf("Cannot connect to db: %s", err)
	}

	runDBMigration(cfg.MigrationURL, cfg.DBSource)

	store := db.NewStore(conn)

	startGrpcServer(cfg, store)
}

func runDBMigration(migrationURL, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Msgf("cannot create a new migrate instance: %s", err)
	}

	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Msgf("failed to run migrate up: %s", err)
	}

	log.Info().Msg("db migrated successfully")
}

func startGrpcServer(cfg util.Config, store db.Store) {
	server, err := gapi.NewServer(cfg, store)
	if err != nil {
		log.Fatal().Msgf("cannot create server: %s", err)
	}

	unaryGrpcLogger := grpc.UnaryInterceptor(gapi.UnaryGRPCLogger)
	serverGrpcLogger := grpc.StreamInterceptor(gapi.ServerGRPCLogger)

	grpcServer := grpc.NewServer(unaryGrpcLogger, serverGrpcLogger)

	pb.RegisterSimpleTODOServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", cfg.GrpcServerAddr)
	if err != nil {
		log.Fatal().Msgf("cannot create listener: %s", err)
	}

	log.Info().Msgf("start gRPC server at %s", listener.Addr())

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal().Msgf("cannot start gRPC server: %s", err)
	}
}
