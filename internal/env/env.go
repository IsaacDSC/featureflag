package env

import (
	"os"
	"sync"
	"time"
)

type Environment struct {
	SecretKey         string
	ServiceClientAT   string
	SDKClientAT       string
	MongoDBURI        string
	MongoDBName       string
	MongoDbIdxTimeout time.Duration
}

var (
	once sync.Once
	env  *Environment
)

func Init() {
	once.Do(func() {
		env = &Environment{
			SecretKey:         os.Getenv("SECRET_KEY"),
			ServiceClientAT:   os.Getenv("SERVICE_CLIENT_AT"),
			SDKClientAT:       os.Getenv("SDK_CLIENT_AT"),
			MongoDBURI:        os.Getenv("MONGODB_URI"),
			MongoDBName:       os.Getenv("MONGODB_NAME"),
			MongoDbIdxTimeout: getDuration("MONGODB_IDX_TIMEOUT", 2*time.Second),
		}
	})
}

func Get() *Environment {
	return env
}

func Override(environment Environment) {
	env = &environment
}

func getDuration(envName string, defaultValue time.Duration) time.Duration {
	envValue := os.Getenv(envName)
	if envValue != "" {
		d, err := time.ParseDuration(envValue)
		if err != nil {
			panic(err)
		}
		defaultValue = d
	}

	return defaultValue
}
