package containers

import (
	"github.com/IsaacDSC/featureflag/internal/audit"
	"github.com/IsaacDSC/featureflag/internal/contenthub"
	"github.com/IsaacDSC/featureflag/internal/featureflag"
	"github.com/IsaacDSC/featureflag/internal/user"
)

const (
	domainFeatureFlag = "feature_flag"
	domainContentHub  = "content_hub"
)

type ServiceContainer struct {
	FeatureFlagService *featureflag.Service
	ContentHubService  *contenthub.Service
	UserService        *user.Service
}

func NewServiceContainer(repositories RepositoryContainer) ServiceContainer {
	featureFlagAuditor := audit.NewAuditService(repositories.AuditRepository, domainFeatureFlag)
	contentHubAuditor := audit.NewAuditService(repositories.AuditRepository, domainContentHub)

	return ServiceContainer{
		FeatureFlagService: featureflag.NewFeatureflagService(repositories.FeatureFlagRepository, featureFlagAuditor),
		ContentHubService:  contenthub.NewContentHubService(repositories.ContentHubRepository, contentHubAuditor),
		UserService:        user.NewUserService(repositories.UserRepository),
	}
}
