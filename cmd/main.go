package main

import (
	"github.com/IsaacDSC/featureflag/internal/containers"
	"github.com/IsaacDSC/featureflag/internal/env"
	"github.com/IsaacDSC/featureflag/internal/handlers"
	"github.com/IsaacDSC/featureflag/internal/middlewares"
	"log"
	"net/http"
	"os"
)

func init() {
	env.Init()
	for i := range env.FilesPaths {
		if _, err := os.ReadFile(env.FilesPaths[i]); err != nil {
			if _, err := os.Create(env.FilesPaths[i]); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func main() {
	repositories := containers.NewRepositoryContainer()
	services := containers.NewServiceContainer(repositories)

	mux := http.NewServeMux()
	handlers := handlers.NewHandlers(services)
	for path, handler := range handlers {
		mux.HandleFunc(path, middlewares.Authorization(handler))
	}

	log.Print("[*] Server started at :3000")
	if err := http.ListenAndServe(":3000", mux); err != nil {
		log.Fatal(err)
	}
}
