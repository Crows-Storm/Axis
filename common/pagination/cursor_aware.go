package pagination

type CursorAware[T any] interface {
	// GetCursorValue returns the string value of the sorting field
	// e.g. time.Time → Unix timestamp string, int → numeric string
	GetCursorValue(field string) string
	// GetID Return record unique identifier
	GetID() string

	ToDomainEdges() []*Edge[T]
}
