package service

import (
	"context"

	"github.com/Crows-Storm/Axis/market/app"
)

func NewApplication(_ context.Context) app.Application {
	return app.Application{}
}
