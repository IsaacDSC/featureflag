package contenthub

type Adapter interface {
	SaveContentHub(input Entity) error
	GetContentHub(key string) (Entity, error)
	GetAllContentHub() (map[string]Entity, error)
	DeleteContentHub(key string) error
}
