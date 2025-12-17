package env

import (
	"log"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Environment struct {
	SecretKey         string        `env:"SECRET_KEY" env-required:"true"`
	ServiceClientAT   string        `env:"SERVICE_CLIENT_AT" env-required:"true"`
	SDKClientAT       string        `env:"SDK_CLIENT_AT" env-required:"true"`
	RepositoryType    string        `env:"REPOSITORY_TYPE" env-default:"jsonfile"`
	MongoDBURI        string        `env:"MONGODB_URI"`
	MongoDBName       string        `env:"MONGODB_NAME"`
	MongoDbIdxTimeout time.Duration `env:"MONGODB_IDX_TIMEOUT" env-default:"2s"`
}

var (
	once sync.Once
	env  *Environment
)

func Init() {
	once.Do(func() {
		env = &Environment{}
		if err := cleanenv.ReadEnv(env); err != nil {
			log.Fatalf("Failed to load environment variables: %v", err)
		}
	})
}

func Get() *Environment {
	return env
}

func Override(environment Environment) {
	env = &environment
}
