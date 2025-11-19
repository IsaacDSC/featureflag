package handlers

import (
	"net/http"

	"github.com/IsaacDSC/featureflag/cmd/containers"
	"github.com/IsaacDSC/featureflag/internal/auth"
	"github.com/IsaacDSC/featureflag/internal/contenthub"
	"github.com/IsaacDSC/featureflag/internal/featureflag"
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
