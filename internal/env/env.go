package env

import (
	"os"
	"sync"
)

type Environment struct {
	SecretKey       string
	ServiceClientAT string
	SDKClientAT     string
	MongoDBURI      string
	MongoDBName     string
}

var (
	once sync.Once
	env  *Environment
)

func Init() {
	once.Do(func() {
		env = &Environment{
			SecretKey:       os.Getenv("SECRET_KEY"),
			ServiceClientAT: os.Getenv("SERVICE_CLIENT_AT"),
			SDKClientAT:     os.Getenv("SDK_CLIENT_AT"),
			MongoDBURI:      os.Getenv("MONGODB_URI"),
			MongoDBName:     os.Getenv("MONGODB_NAME"),
		}
	})
}

func Get() *Environment {
	return env
}

func Override(environment Environment) {
	env = &environment
}
