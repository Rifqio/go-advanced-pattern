package data

import (
	"api.go-rifqio.my.id/internal/validator"
	"math"
	"strings"
)

type Filters struct {
	Page         int    `json:"page"`
	PageSize     int    `json:"pageSize"`
	Sort         string `json:"sort"`
	SortSafeList []string
}

type PaginationMetadata struct {
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	CurrentPage  int `json:"current_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

func calculatePaginationMetadata(totalRecords, page, pageSize int) PaginationMetadata {
	if totalRecords == 0 {
		return PaginationMetadata{}
	}

	return PaginationMetadata{
		PageSize:     pageSize,
		FirstPage:    0,
		CurrentPage:  page,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}

// Check that provided sort field is matches
// one of the entries in SortSafeList[],
// If it does extract the column name otherwise panic
func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafeList {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	panic("unsafe sort parameter: " + f.Sort)
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

func ValidateFilters(validate *validator.Validator, filter *Filters) {
	validate.Check(filter.Page > 0, "page", "CurrentPage is Invalid")
	validate.Check(filter.Page <= 10_000, "page", "CurrentPage Exceed Maximum")

	validate.Check(filter.PageSize <= 50, "page_size", "CurrentPage Size Exceed Maximum")
	validate.Check(filter.PageSize > 0, "page_size", "CurrentPage Size is Invalid")

	validate.Check(validator.In(filter.Sort, filter.SortSafeList...), "sort", "Invalid Sort Value")
}
