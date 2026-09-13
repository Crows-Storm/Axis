package audit

import "context"

type Repository interface {
	Save(context.Context, any) error
}
