package data

import (
	"slices"
	"strings"

	"github.com/arawwad/greenlight/internal/validator"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

type Metadata struct {
	CurrentPage  int
	PageSize     int
	FirstPage    int
	LastPage     int
	TotalRecords int
}

func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be greater than 0")
	v.Check(f.Page < 10_000_000, "page", "must be less than 10 million")

	v.Check(f.PageSize > 0, "page_size", "must be greater than 0")
	v.Check(f.PageSize < 100, "page_size", "must be less than 100")

	v.Check(validator.PermittedValue(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}

func (f Filters) sortColumn() string {
	column := strings.TrimPrefix(f.Sort, "-")
	if slices.Contains(f.SortSafelist, column) {
		return column
	}

	panic("unsafe sort param" + f.Sort)
}

func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}

	return "ASC"
}

func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

func calculateMetadata(page, pageSize, total int) Metadata {
	if total == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     (total + pageSize + 1) / pageSize,
		TotalRecords: total,
	}
}
