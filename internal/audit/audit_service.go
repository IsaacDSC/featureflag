package audit

import (
	"context"
	"time"

	"github.com/IsaacDSC/featureflag/pkg/ctxutils"
	"github.com/IsaacDSC/featureflag/pkg/middlewares"
	"github.com/google/uuid"
)

// Service grava alterações na trilha de auditoria para um domínio fixo
// (ex.: "feature_flag" ou "content_hub"), definido na construção.
type Service struct {
	repo   Adapter
	domain string
}

func NewAuditService(repo Adapter, domain string) *Service {
	return &Service{repo: repo, domain: domain}
}

func (s Service) RecordChange(ctx context.Context, action, project, entityKey string) error {
	emailValue, _ := ctxutils.GetValueCtx(ctx, middlewares.EMAIL_KEY).(string)

	return s.repo.Record(ctx, Entity{
		ID:        uuid.New(),
		Email:     emailValue,
		Action:    action,
		Domain:    s.domain,
		Project:   project,
		EntityKey: entityKey,
		CreatedAt: time.Now(),
	})
}
