package audit

import (
	"context"
	"testing"

	"github.com/IsaacDSC/featureflag/pkg/ctxutils"
	"github.com/IsaacDSC/featureflag/pkg/middlewares"
)

type fakeRepository struct {
	recorded []Entity
}

func (r *fakeRepository) Record(ctx context.Context, entry Entity) error {
	r.recorded = append(r.recorded, entry)
	return nil
}

func TestService_RecordChange(t *testing.T) {
	t.Run("reads the email from the context set by the login middleware", func(t *testing.T) {
		repo := &fakeRepository{}
		svc := NewAuditService(repo, "feature_flag")

		ctx := ctxutils.SetContext(context.Background(), middlewares.EMAIL_KEY, "alice@example.com")
		if err := svc.RecordChange(ctx, ActionUpdated, "checkout", "new-flag"); err != nil {
			t.Fatalf("RecordChange() unexpected error = %v", err)
		}

		if len(repo.recorded) != 1 {
			t.Fatalf("recorded entries = %d, want 1", len(repo.recorded))
		}

		got := repo.recorded[0]
		if got.Email != "alice@example.com" || got.Action != ActionUpdated || got.Domain != "feature_flag" || got.Project != "checkout" || got.EntityKey != "new-flag" {
			t.Errorf("recorded entry = %+v, unexpected fields", got)
		}
	})

	t.Run("no email in context records an empty string instead of panicking", func(t *testing.T) {
		repo := &fakeRepository{}
		svc := NewAuditService(repo, "content_hub")

		if err := svc.RecordChange(context.Background(), ActionDeleted, "", "some-key"); err != nil {
			t.Fatalf("RecordChange() unexpected error = %v", err)
		}

		if got := repo.recorded[0].Email; got != "" {
			t.Errorf("Email = %q, want empty string", got)
		}
	})
}
