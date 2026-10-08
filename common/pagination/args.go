package pagination

func clampLimit(n int) int {
	if n > MaxLimit {
		return MaxLimit
	}
	return n
}

func (a Args) Resolve() (*Resolved, error) {
	if a.First != nil && a.Last != nil {
		return nil, ErrFirstLastConflict
	}
	if a.First != nil && a.Before != nil && *a.Before != "" {
		return nil, ErrInvalidArgs
	}
	if a.Last != nil && a.After != nil && *a.After != "" {
		return nil, ErrInvalidArgs
	}

	if a.First == nil && a.Last == nil {
		limit := DefaultLimit
		return &Resolved{Direction: Forward, Limit: limit, Cursor: a.After}, nil
	}

	if a.First != nil {
		if *a.First < 1 {
			return nil, ErrInvalidArgs
		}
		return &Resolved{
			Direction: Forward,
			Limit:     clampLimit(*a.First),
			Cursor:    a.After,
		}, nil
	}

	if *a.Last < 1 {
		return nil, ErrInvalidArgs
	}
	return &Resolved{
		Direction: Backward,
		Limit:     clampLimit(*a.Last),
		Cursor:    a.Before,
	}, nil
}
