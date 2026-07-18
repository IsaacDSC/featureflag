package audit

import (
	"context"
	"encoding/json"
	"os"

	"github.com/IsaacDSC/featureflag/internal/env"
)

type Repository struct{}

func NewAuditRepository() *Repository {
	return &Repository{}
}

func (r Repository) readStore() ([]Entity, error) {
	b, err := os.ReadFile(env.FilePathAuditLog)
	if err != nil {
		if os.IsNotExist(err) {
			return []Entity{}, nil
		}
		return nil, err
	}

	if len(b) == 0 {
		return []Entity{}, nil
	}

	var store []Entity
	if err := json.Unmarshal(b, &store); err != nil {
		return nil, err
	}

	return store, nil
}

func (r Repository) writeStore(store []Entity) error {
	b, err := json.Marshal(store)
	if err != nil {
		return err
	}

	return os.WriteFile(env.FilePathAuditLog, b, 0644)
}

// Record é sempre um append — não há chave natural para sobrescrever
// (diferente dos demais repositórios jsonfile do projeto).
func (r Repository) Record(ctx context.Context, entry Entity) error {
	store, err := r.readStore()
	if err != nil {
		return err
	}

	store = append(store, entry)

	return r.writeStore(store)
}
