package query

import (
	"context"

	"github.com/Crows-Storm/Axis/common/decorator"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
)

type GetUserStatsQuery struct {
}

type UserStatusAnalysisHandler decorator.QueryHandler[GetUserStatsQuery, map[string]any]

type userStatusAnalysis struct {
	userRepo domain.Repository
}

func NewUserStatusAnalysisHandler(
	repo domain.Repository,
	metricsClient decorator.MetricsClient,
) UserStatusAnalysisHandler {
	if repo == nil {
		panic("nil User Repository")
	}
	return decorator.ApplyQueryDecorators[GetUserStatsQuery, map[string]any](
		userStatusAnalysis{userRepo: repo},
		metricsClient,
	)
}

func (g userStatusAnalysis) Handle(ctx context.Context, query GetUserStatsQuery) (map[string]any, error) {
	info, err := g.userRepo.GetStats(ctx)
	if err != nil {
		return nil, err
	}
	return info, err
}
