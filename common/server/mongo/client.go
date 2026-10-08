package mongo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

const (
	// TimestampField is the time-series time field. Documents must store a BSON datetime here.
	TimestampField = "timestamp"
	// MetadataField groups related series. Keep it low-cardinality (service + event name).
	MetadataField = "metadata"
	// GranularitySeconds matches audit events, which arrive per command rather than per sensor tick.
	GranularitySeconds = "seconds"
)

// Client is a connected MongoDB handle. The URI is never logged.
type Client struct {
	raw *mongo.Client
	db  *mongo.Database
}

// Connect pings the deployment and returns a client for database.
func Connect(ctx context.Context, uri, database string) (*Client, error) {
	if uri == "" {
		return nil, fmt.Errorf("mongo uri is empty")
	}
	if database == "" {
		return nil, fmt.Errorf("mongo database is empty")
	}

	raw, err := mongo.Connect(options.Client().ApplyURI(uri).SetServerSelectionTimeout(5 * time.Second))
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := raw.Ping(pingCtx, readpref.Primary()); err != nil {
		_ = raw.Disconnect(context.Background())
		return nil, fmt.Errorf("mongo ping: %w", err)
	}

	return &Client{raw: raw, db: raw.Database(database)}, nil
}

// Database returns the configured database.
func (c *Client) Database() *mongo.Database {
	if c == nil {
		return nil
	}
	return c.db
}

// Collection returns a handle in the configured database.
func (c *Client) Collection(name string) *mongo.Collection {
	if c == nil || c.db == nil {
		return nil
	}
	return c.db.Collection(name)
}

// Close disconnects the underlying client.
func (c *Client) Close(ctx context.Context) error {
	if c == nil || c.raw == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return c.raw.Disconnect(ctx)
}

// Disconnect closes the underlying client.
func (c *Client) Disconnect(ctx context.Context) error {
	if c == nil || c.raw == nil {
		return nil
	}
	if err := c.raw.Disconnect(ctx); err != nil {
		return fmt.Errorf("mongo disconnect: %w", err)
	}
	return nil
}

// InitCollection creates each missing collection and ensures indexes.
// Time-series documents use timestamp, metadata, and seconds granularity.
// An existing collection is left in place; indexes are still created when missing.
func (c *Client) InitCollection(ctx context.Context, docs ...Document) error {
	if c == nil || c.db == nil {
		return fmt.Errorf("mongo database is nil")
	}
	unique, err := collectionsToCreate(docs, nil)
	if err != nil || len(unique) == 0 {
		return err
	}
	for _, doc := range unique {
		if _, err := indexModelsOf(doc); err != nil {
			return err
		}
	}

	names := make([]string, len(unique))
	for i, doc := range unique {
		names[i] = strings.TrimSpace(doc.CollectionName())
	}
	existing, err := c.db.ListCollectionNames(ctx, bson.D{{
		Key:   "name",
		Value: bson.D{{Key: "$in", Value: names}},
	}})
	if err != nil {
		return fmt.Errorf("list collections: %w", err)
	}
	pending, err := collectionsToCreate(docs, existing)
	if err != nil {
		return err
	}

	for _, doc := range pending {
		name := strings.TrimSpace(doc.CollectionName())
		if err := createCollection(ctx, c.db, name, doc.IsTimeSeries()); err != nil {
			return err
		}
	}
	for _, doc := range unique {
		if err := ensureIndexes(ctx, c.db, doc); err != nil {
			return err
		}
	}
	return nil
}

func createCollection(ctx context.Context, db *mongo.Database, name string, timeSeries bool) error {
	var err error
	kind := "collection"
	if timeSeries {
		kind = "time series collection"
		err = db.CreateCollection(ctx, name, options.CreateCollection().SetTimeSeriesOptions(
			options.TimeSeries().
				SetTimeField(TimestampField).
				SetMetaField(MetadataField).
				SetGranularity(GranularitySeconds),
		))
	} else {
		err = db.CreateCollection(ctx, name)
	}
	if err == nil || isNamespaceExists(err) {
		return nil
	}
	return fmt.Errorf("create %s %s: %w", kind, name, err)
}

// collectionsToCreate validates documents and drops ones that already exist.
// Duplicate names are kept once. The same name cannot be both normal and time-series.
func collectionsToCreate(docs []Document, existing []string) ([]Document, error) {
	have := make(map[string]struct{}, len(existing))
	for _, name := range existing {
		have[name] = struct{}{}
	}

	seen := make(map[string]bool, len(docs))
	pending := make([]Document, 0, len(docs))
	for _, doc := range docs {
		if doc == nil {
			return nil, fmt.Errorf("mongo document is nil")
		}
		name := strings.TrimSpace(doc.CollectionName())
		if name == "" {
			return nil, fmt.Errorf("mongo collection name is empty")
		}
		if timeSeries, ok := seen[name]; ok {
			if timeSeries != doc.IsTimeSeries() {
				return nil, fmt.Errorf("collection %s has conflicting types", name)
			}
			continue
		}
		seen[name] = doc.IsTimeSeries()
		if _, ok := have[name]; ok {
			continue
		}
		pending = append(pending, doc)
	}
	return pending, nil
}

func isNamespaceExists(err error) bool {
	var cmd mongo.CommandError
	if errors.As(err, &cmd) {
		return cmd.Code == 48 || cmd.Name == "NamespaceExists"
	}
	return false
}
