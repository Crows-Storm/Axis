package principal

import (
	"context"

	"github.com/Crows-Storm/Axis/common/domain/principal"
)

type Repository interface {
	GetPrincipal(ctx context.Context, token string) (*principal.Principal, error)
}
