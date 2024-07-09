package interfaces

import "github.com/IsaacDSC/featureflag/internal/domain/entity"

type FeatureFlagRepository interface {
	SaveFF(input entity.Featureflag) error
	GetAllFF() (map[string]entity.Featureflag, error)
	GetFF(key string) (entity.Featureflag, error)
	DeleteFF(key string) error
}
