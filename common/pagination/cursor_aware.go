package pagination

type CursorAware interface {
	GetCursorValue(field string) string
	GetID() string
}
