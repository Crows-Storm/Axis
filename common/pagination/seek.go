package pagination

import "strings"

type QueryConfig struct {
	SortField string
	SortOrder string
	IDField   string
	Count     bool
}

func DefaultConfig() QueryConfig {
	return QueryConfig{
		SortField: "CreatedTime",
		SortOrder: "ASC",
		IDField:   "ID",
		Count:     true,
	}
}

func NewestFirst(sortField, idField string) QueryConfig {
	return QueryConfig{
		SortField: sortField,
		SortOrder: "DESC",
		IDField:   idField,
	}
}

func (c QueryConfig) sortAscending() bool {
	return !strings.EqualFold(strings.TrimSpace(c.SortOrder), "DESC")
}

func (c QueryConfig) QueryAscending(dir Direction) bool {
	asc := c.sortAscending()
	if dir == Backward {
		return !asc
	}
	return asc
}

func (c QueryConfig) OrderSQL(dir Direction) string {
	order := "ASC"
	if !c.QueryAscending(dir) {
		order = "DESC"
	}
	return order
}

type Seek struct {
	Value   string
	ID      string
	Greater bool
}

func NewSeek(resolved *Resolved, cfg QueryConfig) (*Seek, error) {
	if resolved == nil || resolved.Cursor == nil || *resolved.Cursor == "" {
		return nil, nil
	}
	value, id, err := DecodeCursor(*resolved.Cursor)
	if err != nil {
		return nil, err
	}
	return &Seek{
		Value:   value,
		ID:      id,
		Greater: cfg.QueryAscending(resolved.Direction),
	}, nil
}
