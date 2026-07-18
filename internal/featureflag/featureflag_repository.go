package featureflag

import (
	"context"
	"encoding/json"
	"os"
	"sort"

	"github.com/IsaacDSC/featureflag/internal/env"
	"github.com/IsaacDSC/featureflag/pkg/errorutils"
)

type Repository struct{}

func NewFeatureFlagRepository() *Repository {
	return &Repository{}
}

// projectStore representa o layout do arquivo: project -> flagName -> Entity.
type projectStore map[string]map[string]Entity

func (fr Repository) readStore() (projectStore, error) {
	b, err := os.ReadFile(env.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return projectStore{}, nil
		}
		return nil, err
	}

	if len(b) == 0 {
		return projectStore{}, nil
	}

	var store projectStore
	if err := json.Unmarshal(b, &store); err != nil {
		return nil, err
	}

	if store == nil {
		store = projectStore{}
	}

	return store, nil
}

func (fr Repository) writeStore(store projectStore) error {
	b, err := json.Marshal(store)
	if err != nil {
		return err
	}

	return os.WriteFile(env.FilePath, b, 0644)
}

func (fr Repository) SaveFF(ctx context.Context, project string, input Entity) error {
	store, err := fr.readStore()
	if err != nil {
		return err
	}

	if store[project] == nil {
		store[project] = map[string]Entity{}
	}

	store[project][input.FlagName] = input

	return fr.writeStore(store)
}

func (fr Repository) GetFF(ctx context.Context, project, key string) (Entity, error) {
	store, err := fr.readStore()
	if err != nil {
		return Entity{}, err
	}

	flags, ok := store[project]
	if !ok {
		return Entity{}, errorutils.NewNotFoundError("featureflag")
	}

	if output, ok := flags[key]; ok {
		return output, nil
	}

	return Entity{}, errorutils.NewNotFoundError("featureflag")
}

func (fr Repository) GetAllFF(ctx context.Context, project string) (map[string]Entity, error) {
	store, err := fr.readStore()
	if err != nil {
		return map[string]Entity{}, err
	}

	if flags, ok := store[project]; ok {
		return flags, nil
	}

	return map[string]Entity{}, nil
}

func (fr Repository) ListProjects(ctx context.Context) ([]string, error) {
	store, err := fr.readStore()
	if err != nil {
		return nil, err
	}

	projects := make([]string, 0, len(store))
	for project := range store {
		projects = append(projects, project)
	}

	sort.Strings(projects)

	return projects, nil
}

func (fr Repository) DeleteFF(ctx context.Context, project, key string) error {
	store, err := fr.readStore()
	if err != nil {
		return err
	}

	if flags, ok := store[project]; ok {
		delete(flags, key)
	}

	return fr.writeStore(store)
}
