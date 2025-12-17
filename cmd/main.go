package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IsaacDSC/featureflag/cmd/containers"
	"github.com/IsaacDSC/featureflag/internal/env"
	"github.com/IsaacDSC/featureflag/pkg/handlers"
	"github.com/IsaacDSC/featureflag/pkg/middlewares"
	"github.com/IsaacDSC/featureflag/pkg/pubsub"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var rdb *redis.Client

func init() {
	env.Init()
	for i := range env.FilesPaths {
		if _, err := os.ReadFile(env.FilesPaths[i]); err != nil {
			if _, err := os.Create(env.FilesPaths[i]); err != nil {
				log.Fatal(err)
			}
		}
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rdb = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
}

func main() {
	environment := env.Get()

	var repositories containers.RepositoryContainer
	if environment.RepositoryType == "jsonfile" {
		repositories = containers.NewRepositoryContainer()
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		clientOptions := options.Client().ApplyURI(environment.MongoDBURI)
		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			log.Fatalf("Failed to connect to MongoDB: %v", err)
		}

		// Ping MongoDB to verify connection
		if err := client.Ping(ctx, nil); err != nil {
			log.Fatalf("Failed to ping MongoDB: %v", err)
		}

		defer client.Disconnect(ctx)

		repositories = containers.NewRepositoryContainerMongodb(client, environment.MongoDBName)
	}

	pub := pubsub.NewPublisher(rdb)
	services := containers.NewServiceContainer(repositories, pub)
	sub := pubsub.NewSubscriber(rdb)

	mux := http.NewServeMux()
	handlers := handlers.NewHandlers(services, sub)
	for path, handler := range handlers {
		// mux.HandleFunc(path, middlewares.Authorization(handler))
		mux.HandleFunc(path, middlewares.Logger(handler))
	}

	server := &http.Server{
		Addr:    ":3000",
		Handler: mux,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Print("[*] Server started at :3000")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stop
	log.Print("\n[*] Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	if err := rdb.Close(); err != nil {
		log.Printf("Error closing Redis connection: %v", err)
	}

	log.Print("[*] Server stopped")
}
