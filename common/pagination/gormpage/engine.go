package gormpage

import (
	"fmt"

	"github.com/Crows-Storm/Axis/common/pagination"
	"gorm.io/gorm"
)

func FetchConnection[T pagination.CursorAware](
	_ *gorm.DB,
	model *gorm.DB,
	config pagination.QueryConfig,
	args *pagination.Args,
) (*pagination.Connection[T], error) {
	if args == nil {
		args = &pagination.Args{}
	}
	resolved, err := args.Resolve()
	if err != nil {
		return nil, err
	}
	if config.SortField == "" || config.IDField == "" {
		return nil, fmt.Errorf("%w: sort fields required", pagination.ErrInvalidArgs)
	}

	var totalCount int64
	if config.Count {
		if err := model.Count(&totalCount).Error; err != nil {
			return nil, fmt.Errorf("count failed: %w", err)
		}
	}

	query := model.Session(&gorm.Session{NewDB: false})
	seek, err := pagination.NewSeek(resolved, config)
	if err != nil {
		return nil, err
	}
	if seek != nil {
		cmp := "<"
		if seek.Greater {
			cmp = ">"
		}
		query = query.Where(
			fmt.Sprintf("(%s %s ?) OR (%s = ? AND %s %s ?)", config.SortField, cmp, config.SortField, config.IDField, cmp),
			seek.Value, seek.Value, seek.ID,
		)
	}

	order := config.OrderSQL(resolved.Direction)
	query = query.
		Order(fmt.Sprintf("%s %s, %s %s", config.SortField, order, config.IDField, order)).
		Limit(resolved.Limit + 1)

	var results []T
	if err := query.Find(&results).Error; err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	return pagination.Wrap(results, resolved, config, totalCount), nil
}
