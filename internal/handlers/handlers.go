package handlers

import (
	"github.com/IsaacDSC/featureflag/internal/containers"
	"github.com/IsaacDSC/featureflag/internal/domains/auth"
	"github.com/IsaacDSC/featureflag/internal/domains/contenthub"
	"github.com/IsaacDSC/featureflag/internal/domains/featureflag"
	"net/http"
)

func NewHandlers(services containers.ServiceContainer) map[string]func(w http.ResponseWriter, r *http.Request) {
	output := make(map[string]func(w http.ResponseWriter, r *http.Request))

	for k, v := range auth.NewAuthHandler().GetRoutes() {
		output[k] = v
	}

	for k, v := range featureflag.NewFeatureFlagHandler(services.FeatureFlagService).GetRoutes() {
		output[k] = v
	}

	for k, v := range contenthub.NewContenthubHandler(services.ContentHubService).GetRoutes() {
		output[k] = v
	}

	return output
}
