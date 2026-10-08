package adapters

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Crows-Storm/Axis/audit/domain/audit"
	"github.com/Crows-Storm/Axis/common/pagination"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var rootPageConfig = pagination.NewestFirst(fieldTimestamp, fieldEventID)

type RecordMongoRepository struct {
	col *mongo.Collection
}

func NewRecordMongoRepository(col *mongo.Collection) *RecordMongoRepository {
	if col == nil {
		panic("nil audit collection")
	}
	return &RecordMongoRepository{col: col}
}

func (r *RecordMongoRepository) Append(ctx context.Context, rec *audit.AuditEvent) error {
	if rec == nil {
		return audit.ErrInvalidRecord
	}
	if _, err := r.col.InsertOne(ctx, rec); err != nil {
		return fmt.Errorf("append audit event: %w", err)
	}
	return nil
}

func (r *RecordMongoRepository) GetByEventID(ctx context.Context, eventID string) (*audit.AuditEvent, error) {
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return nil, audit.ErrInvalidRecord
	}
	var rec audit.AuditEvent
	err := r.col.FindOne(ctx, bson.D{{Key: fieldEventID, Value: eventID}}).Decode(&rec)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, audit.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get audit event: %w", err)
	}
	return &rec, nil
}

func (r *RecordMongoRepository) PageRoots(ctx context.Context, filter audit.RootFilter, args pagination.Args) (*pagination.Connection[*audit.AuditEvent], error) {
	resolved, err := args.Resolve()
	if err != nil {
		return nil, err
	}
	seek, err := pagination.NewSeek(resolved, rootPageConfig)
	if err != nil {
		return nil, err
	}
	query, err := rootQuery(filter, seek)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", pagination.ErrInvalidCursor, err)
	}

	opts := options.Find().
		SetSort(pagination.MongoSort(fieldTimestamp, fieldEventID, rootPageConfig.QueryAscending(resolved.Direction))).
		SetLimit(int64(resolved.Limit + 1))
	cursor, err := r.col.Find(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("page root events: %w", err)
	}
	defer func() { _ = cursor.Close(ctx) }()

	rows := make([]*audit.AuditEvent, 0)
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, fmt.Errorf("decode root events: %w", err)
	}
	return pagination.Wrap(rows, resolved, rootPageConfig, 0), nil
}

func (r *RecordMongoRepository) ListByCorrelation(ctx context.Context, correlationID string) ([]*audit.AuditEvent, error) {
	correlationID = strings.TrimSpace(correlationID)
	if correlationID == "" {
		return []*audit.AuditEvent{}, nil
	}
	opts := options.Find().
		SetSort(pagination.MongoSort(fieldTimestamp, fieldEventID, true)).
		SetLimit(int64(audit.MaxGraphEvents + 1))
	cursor, err := r.col.Find(ctx, bson.D{{Key: fieldCorrelationID, Value: correlationID}}, opts)
	if err != nil {
		return nil, fmt.Errorf("list by correlation: %w", err)
	}
	defer func() { _ = cursor.Close(ctx) }()

	rows := make([]*audit.AuditEvent, 0)
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, fmt.Errorf("decode correlation events: %w", err)
	}
	if len(rows) > audit.MaxGraphEvents {
		rows = rows[:audit.MaxGraphEvents]
	}
	return rows, nil
}
