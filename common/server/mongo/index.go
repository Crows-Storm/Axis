package mongo

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// IndexModels builds create-index models. Field names are trimmed. ASC is 1 and DESC is -1.
func IndexModels(specs []*IndexInfo) ([]mongo.IndexModel, error) {
	seen := make(map[string]struct{}, len(specs))
	models := make([]mongo.IndexModel, 0, len(specs))
	for _, spec := range specs {
		if spec == nil {
			return nil, fmt.Errorf("mongo index spec is nil")
		}
		name := strings.TrimSpace(spec.Name)
		if name == "" || len(spec.Fields) == 0 {
			return nil, fmt.Errorf("mongo index spec is invalid")
		}
		if _, ok := seen[name]; ok {
			return nil, fmt.Errorf("mongo index %s is duplicated", name)
		}
		seen[name] = struct{}{}

		direction, err := indexDirection(spec.Type)
		if err != nil {
			return nil, fmt.Errorf("mongo index %s: %w", name, err)
		}
		keys := make(bson.D, 0, len(spec.Fields))
		for _, field := range spec.Fields {
			field = strings.TrimSpace(field)
			if field == "" {
				return nil, fmt.Errorf("mongo index %s has an empty field", name)
			}
			keys = append(keys, bson.E{Key: field, Value: direction})
		}
		models = append(models, mongo.IndexModel{
			Keys:    keys,
			Options: options.Index().SetName(name),
		})
	}
	return models, nil
}

func indexDirection(kind IndexType) (int32, error) {
	switch kind {
	case "", ASC:
		return 1, nil
	case DESC:
		return -1, nil
	default:
		return 0, fmt.Errorf("unsupported index type %q", kind)
	}
}

func indexModelsOf(doc Document) ([]mongo.IndexModel, error) {
	indexed, ok := doc.(CollectionIndex)
	if !ok || indexed == nil {
		return nil, nil
	}
	return IndexModels(indexed.Index())
}

func ensureIndexes(ctx context.Context, db *mongo.Database, doc Document) error {
	models, err := indexModelsOf(doc)
	if err != nil || len(models) == 0 {
		return err
	}
	name := strings.TrimSpace(doc.CollectionName())
	if _, err := db.Collection(name).Indexes().CreateMany(ctx, models); err != nil {
		return fmt.Errorf("ensure indexes %s: %w", name, err)
	}
	return nil
}
