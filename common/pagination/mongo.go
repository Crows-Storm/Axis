package pagination

import "go.mongodb.org/mongo-driver/v2/bson"

func MongoSeekFilter(sortField, idField string, seek *Seek, sortValue any) bson.D {
	if seek == nil {
		return nil
	}
	op := "$lt"
	if seek.Greater {
		op = "$gt"
	}
	return bson.D{{
		Key: "$or",
		Value: bson.A{
			bson.D{{Key: sortField, Value: bson.D{{Key: op, Value: sortValue}}}},
			bson.D{
				{Key: sortField, Value: sortValue},
				{Key: idField, Value: bson.D{{Key: op, Value: seek.ID}}},
			},
		},
	}}
}

func MongoSort(sortField, idField string, ascending bool) bson.D {
	dir := int32(-1)
	if ascending {
		dir = 1
	}
	return bson.D{
		{Key: sortField, Value: dir},
		{Key: idField, Value: dir},
	}
}
