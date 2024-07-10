package env

import (
	"os"
	"sync"
)

type Environment struct {
	SecretKey       string
	ServiceClientAT string
	SDKClientAT     string
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
		}
	})
}

func Get() *Environment {
	return env
}

func Override(environment Environment) {
	env = &environment
}
