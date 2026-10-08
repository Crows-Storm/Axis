package service

import (
	"context"

	"github.com/Crows-Storm/Axis/audit/adapters"
	"github.com/Crows-Storm/Axis/audit/app"
	"github.com/Crows-Storm/Axis/audit/app/command"
	"github.com/Crows-Storm/Axis/audit/app/query"
	domain "github.com/Crows-Storm/Axis/audit/domain/audit"
	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/server/metrics"
	"github.com/Crows-Storm/Axis/common/server/mongo"
)

type ApplicationDependencies struct {
	MongoClient *mongo.Client
}

func NewApplication(ctx context.Context, deps ApplicationDependencies) app.Application {
	// Init all time series and document collection
	names := make([]mongo.Document, 0)
	audit := new(domain.AuditEvent)
	names = append(names, audit)
	if err := deps.MongoClient.InitCollection(ctx, names...); err != nil {
		logger.Fatalf("audit time series init failed: %v", err)
	}

	repo := adapters.NewRecordMongoRepository(deps.MongoClient.Collection(audit.CollectionName()))
	return newApplication(ctx, repo, metrics.TodoMetrics{})
}

func newApplication(_ context.Context, repo domain.Repository, metricsClient decorator.MetricsClient) app.Application {
	return app.Application{
		Commands: app.Commands{
			RecordEvent: command.NewRecordAuditEventCommandHandler(repo, metricsClient),
		},
		Queries: app.Queries{
			PageRoots:     query.NewPageRootsQueryHandler(repo, metricsClient),
			GetEventGraph: query.NewGetEventGraphQueryHandler(repo, metricsClient),
		},
	}
}
