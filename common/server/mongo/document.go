package mongo

type Document interface {
	CollectionName() string // collection name
	IsTimeSeries() bool     // is time series collection
}

// The CollectionIndex is definition index
// struct to implement this interface, definition self the index combination
type CollectionIndex interface {
	Index() []*IndexInfo
}

type IndexInfo struct {
	Name   string
	Fields []string

	Type IndexType
}

type IndexType string

const (
	ASC  IndexType = "1"
	DESC IndexType = "-1"
	//DSPHERE IndexType = "2dsphere"
	//TEXT    IndexType = "text"
)
