package adapters

import (
	"strings"
	"time"

	"github.com/Crows-Storm/Axis/audit/domain/audit"
	"github.com/Crows-Storm/Axis/common/pagination"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	fieldTimestamp     = "timestamp"
	fieldEventID       = "event_id"
	fieldEventName     = "event_name"
	fieldService       = "metadata.service"
	fieldCorrelationID = "metadata.correlation_id"
	fieldCausationID   = "metadata.causation_id"
)

func rootQuery(filter audit.RootFilter, seek *pagination.Seek) (bson.D, error) {
	query := bson.D{
		{Key: fieldCorrelationID, Value: bson.D{{Key: "$gt", Value: ""}}},
		{Key: fieldCausationID, Value: bson.D{{Key: "$not", Value: bson.D{{Key: "$gt", Value: ""}}}}},
	}
	if service := strings.TrimSpace(filter.Service); service != "" {
		query = append(query, bson.E{Key: fieldService, Value: service})
	}
	if name := strings.TrimSpace(filter.EventName); name != "" {
		query = append(query, bson.E{Key: fieldEventName, Value: name})
	}
	if !filter.From.IsZero() || !filter.To.IsZero() {
		bound := bson.D{}
		if !filter.From.IsZero() {
			bound = append(bound, bson.E{Key: "$gte", Value: filter.From.UTC()})
		}
		if !filter.To.IsZero() {
			bound = append(bound, bson.E{Key: "$lt", Value: filter.To.UTC()})
		}
		query = append(query, bson.E{Key: fieldTimestamp, Value: bound})
	}
	if seek != nil {
		occurred, err := parseCursorTime(seek.Value)
		if err != nil {
			return nil, err
		}
		seekFilter := pagination.MongoSeekFilter(fieldTimestamp, fieldEventID, seek, occurred)
		query = append(query, seekFilter...)
	}
	return query, nil
}

func parseCursorTime(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return t.UTC(), nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}
