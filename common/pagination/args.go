package pagination

import "errors"

var (
	ErrInvalidArgs       = errors.New("invalid pagination arguments")
	ErrFirstLastConflict = errors.New("cannot use both 'first' and 'last'")
)

// Resolve resolves external Args into internal Resolved parameters.
func (a *Args) Resolve() (*Resolved, error) {
	if a.First != nil && a.Last != nil {
		return nil, ErrFirstLastConflict
	}

	if a.First == nil && a.Last == nil {
		// The default forward option retrieves the DefaultLimit line.
		defaultLimit := DefaultLimit
		a.First = &defaultLimit
	}

	if a.First != nil {
		limit := *a.First
		if limit > MaxLimit {
			limit = MaxLimit
		}
		return &Resolved{
			Direction: Forward,
			Limit:     limit,
			Cursor:    a.After,
		}, nil
	}

	// Last + Before
	limit := *a.Last
	if limit > MaxLimit {
		limit = MaxLimit
	}
	return &Resolved{
		Direction: Backward,
		Limit:     limit,
		Cursor:    a.Before,
	}, nil
}
