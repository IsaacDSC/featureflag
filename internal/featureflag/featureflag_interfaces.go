package featureflag

type Adapter interface {
	SaveFF(input Entity) error
	GetAllFF() (map[string]Entity, error)
	GetFF(key string) (Entity, error)
	DeleteFF(key string) error
}
