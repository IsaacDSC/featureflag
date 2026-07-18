package featureflag

import "context"

// noopAuditor satisfaz Auditor sem gravar nada — usado nos testes que não
// exercitam a trilha de auditoria diretamente.
type noopAuditor struct{}

func (noopAuditor) RecordChange(ctx context.Context, action, project, entityKey string) error {
	return nil
}
