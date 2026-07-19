package audit

import "context"

type Adapter interface {
	Record(ctx context.Context, entry Entity) error
}
