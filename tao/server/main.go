package main

import (
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

var addr = "0.0.0.0:70077"

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
		pgPort = ""
	}

	db := pg.Connect(&pg.Options{
		User:     pgUsername,
		Addr:     fmt.Sprintf("%s:%s", pgHost, pgPort),
		Password: pgPassword,
		Database: pgDBName,
	})

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
				Temp: true,
			})
			if err != nil {
				return err
			}
		}
		return nil
	}

	err = createSchema(db)

	if err != nil {
		panic("Failed to create schema")
	}

	pb.RegisterTaoServiceServer(s, &Server{
		redisClient: rdb,
		pgDB:        db,
	})

	if err = s.Serve(lis); err != nil {
		log.Fatal("Failed to serve: ", err)
	}
}
