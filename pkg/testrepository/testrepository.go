package testrepository

import (
	"os"
	"testing"
)

type SetupRepositoryTest struct {
	contenthubPath string
}

func NewSetupRepositoryTest(contenthubPath string) *SetupRepositoryTest {
	return &SetupRepositoryTest{contenthubPath: contenthubPath}
}

func (s *SetupRepositoryTest) Setup() {
	os.WriteFile(s.contenthubPath, []byte("{}"), 0644)
}

func (s *SetupRepositoryTest) TearDown() {
	os.Remove(s.contenthubPath)
}

type ConfigTestRepository struct {
	TearDown bool
	TearUp   bool
}

func (s *SetupRepositoryTest) Run(conf ConfigTestRepository, t *testing.T, name string, fn func(t *testing.T)) (output bool) {
	if conf.TearUp {
		s.Setup()
	}

	output = t.Run(name, fn)

	if conf.TearDown {
		s.TearDown()
	}

	return
}
