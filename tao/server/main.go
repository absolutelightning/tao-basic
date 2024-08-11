package main

import (
	"context"
	"fmt"
	"github.com/go-pg/pg/v10/orm"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"

	pb "github.com/absolutelightning/tao-basic/tao/proto"
	"github.com/go-pg/pg/v10"
)

var addr string = "localhost:7051"

type Server struct {
	pb.TaoServiceServer
	redisClient *redis.Client
	pgDB        *pg.DB
}

func main() {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("Failed to listen on address: ", addr)
	}
	log.Println("Listening on address: ", addr)

	s := grpc.NewServer()

	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	pgHost := os.Getenv("PG_HOST")
	pgPort := os.Getenv("PG_PORT")
	pgUsername := os.Getenv("PG_USERNAME")
	pgPassword := os.Getenv("PG_PASSWORD")
	pgDBName := os.Getenv("PG_DBNAME")

	if redisHost == "" {
		redisHost = "localhost"
	}

	if redisPort == "" {
		redisPort = "6379"
	}

	if pgHost == "" {
		pgHost = "localhost"
	}

	if pgPort == "" {
		pgPort = "5432"
	}

	db := pg.Connect(&pg.Options{
		User:     pgUsername,
		Addr:     fmt.Sprintf("%s:%s", pgHost, pgPort),
		Password: pgPassword,
		Database: pgDBName,
	})

	err = db.Ping(context.Background())
	if err != nil {
		panic(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// createSchema creates database schema for User and Story models.
	createSchema := func(db *pg.DB) error {
		models := []interface{}{
			(*Object)(nil),
			(*Association)(nil),
		}

		for _, model := range models {
			err := db.Model(model).CreateTable(&orm.CreateTableOptions{
				Temp:        false,
				IfNotExists: true,
			})
			if err != nil {
				return err
			}
		}
		return nil
	}

	err = createSchema(db)
	if err != nil {
		panic(fmt.Sprintf("Failed to create schema %s", err.Error()))
	}

	createIndexes := func(db *pg.DB) error {
		_, err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_jsonb_data ON objects USING gin (data);
		CREATE INDEX IF NOT EXISTS idx_jsonb_btree_name ON objects ((data->>'name'));
		CREATE INDEX IF NOT EXISTS idx_jsonb_btree_version ON objects ((data->>'version'));
	`)
		return err
	}

	err = createIndexes(db)
	if err != nil {
		panic(fmt.Sprintf("Failed to create schema %s", err.Error()))
	}

	pb.RegisterTaoServiceServer(s, &Server{
		redisClient: rdb,
		pgDB:        db,
	})

	if err = s.Serve(lis); err != nil {
		log.Fatal("Failed to serve: ", err)
	}
}
