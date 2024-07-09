package main

import (
	"github.com/IsaacDSC/featureflag/internal/domain"
	"github.com/IsaacDSC/featureflag/internal/infra"
	"github.com/IsaacDSC/featureflag/internal/web"
	"log"
	"net/http"
	"os"
)

func init() {
	if _, err := os.ReadFile(infra.FilePath); err != nil {
		if _, err := os.Create(infra.FilePath); err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	repository := infra.NewFeatureFlagRepository()
	service := domain.NewFeatureflagService(repository)

	mux := http.NewServeMux()
	handlers := web.NewHandler(service).GetRoutes()
	for router, handler := range handlers {
		mux.HandleFunc(router, web.ApplicationJSON(handler))
	}

	if err := http.ListenAndServe(":3000", mux); err != nil {
		log.Fatal(err)
	}
}
