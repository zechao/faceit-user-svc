// Package query define struct used for pagination
package query

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/zechao/faceit-user-svc/errors"
	"gorm.io/gorm"
)

const (
	// defalutPageSize defines the default number of items that can be returned per page.
	defalutPageSize = 100

	defaultSortOrder = "desc"
	defaultSortBy    = "created_at"

	paramPage      = "page"
	paramPageSize  = "page_size"
	paramSortOrder = "sort_order"
	paramSortBy    = "sort_by"

	ErrCodeInvalidParameter = "INVALID_QUERY_PARAMETERS"
)

// supportedFilter is a map of supported filter parameters, boolean value is to indicate if its filter
var supportedQuery = map[string]bool{
	"first_name": true,
	"last_name":  true,
	"nick_name":  true,
	"password":   true,
	"email":      true,
	"country":    true,
	"created_at": true,
	"updated_at": true,
	"id":         true,
	"page":       false,
	"page_size":  false,
	"sort_order": false,
	"sort_by":    false,
}

// Query represents the parsed query parameters for pagination, sorting and filtering.
type Query struct {
	Page     int
	PageSize int
	// SortOrder default value is desc
	SortOrder string
	// sort by specific field, default created_at
	SortBy string
	// Allow multiple values per filter
	Filters map[string][]string
}

// PaginationResponse is a generic struct that holds the paginated data and metadata.
type PaginationResponse[T any] struct {
	Page         int                 `json:"page"`
	PageSize     int                 `json:"page_size"`
	TotalRecords int64               `json:"total_records"`
	SortBy       string              `json:"sort_by"`
	SortOrder    string              `json:"sort_order"`
	Filters      map[string][]string `json:"filters"`
	Data         []T                 `json:"data"`
}

// QueryFromURL parses the query parameters from a URL and returns a Query object.
// It validates the page, page_size, sort_order, and sort_by parameters.
// If any parameter is invalid, it returns an error with details.
// be aware we are not restricting the max number of page_size here.
// the filter parameters here so that we can have flexible query options.
// for example, /users?name=John&country=UK&country=ES
// will be parsed as Filters: {"name": ["John"], "country": ["ES", "UK"]}
func QueryFromURL(params url.Values) (*Query, error) {
	q := Query{
		Page:      1,
		PageSize:  defalutPageSize,
		SortOrder: defaultSortOrder,
		SortBy:    defaultSortBy,
		Filters:   make(map[string][]string),
	}

	var details []errors.Detail

	if val := params.Get(paramPage); val != "" {
		page, err := strconv.Atoi(val)
		if page > 1 {
			q.Page = page
		}
		if err != nil {
			details = append(details, errors.Detail{
				Field:       paramPage,
				Description: "page must be a number",
			})
		}
		if err == nil && page < 1 {
			details = append(details, errors.Detail{
				Field:       paramPage,
				Description: "page number must be greater than 0",
			})
		}
	}

	if val := params.Get(paramPageSize); val != "" {

		pageSize, err := strconv.Atoi(val)
		if pageSize > 0 {
			q.PageSize = pageSize
		}
		if err != nil {
			details = append(details, errors.Detail{
				Field:       paramPageSize,
				Description: "page_size must be a number",
			})
		}
		if err == nil && pageSize < 1 {
			details = append(details, errors.Detail{
				Field:       paramPageSize,
				Description: "page_size must be greater than 0",
			})
		}
	}

	q.SortBy = params.Get(paramSortBy)
	if q.SortBy == "" {
		q.SortBy = defaultSortBy
	}

	sortOrder := params.Get(paramSortOrder)
	if sortOrder == "" {
		q.SortOrder = defaultSortOrder
	}

	if sortOrder != "" {
		if !strings.EqualFold(sortOrder, "desc") && !strings.EqualFold(sortOrder, "asc") {
			details = append(details, errors.Detail{
				Field:       paramSortOrder,
				Description: "sort_order must be desc or asc",
			})
		}
		q.SortOrder = sortOrder
	}

	// check filters
	for key, values := range params {
		isFilter, validQuery := supportedQuery[key]
		if !validQuery {
			details = append(details, errors.Detail{
				Field:       key,
				Description: fmt.Sprintf("parameter %s is not supported", key),
			})
		}
		if isFilter {
			q.Filters[key] = values
		}
	}

	if len(details) != 0 {
		return nil, errors.NewWrongInput(ErrCodeInvalidParameter, details...)
	}

	return &q, nil
}

// ApplyQuery applies the query parameters to a GORM database query.
func (query Query) ApplyQuery(db *gorm.DB) *gorm.DB {
	// Default values
	if query.SortBy == "" {
		query.SortBy = "created_at"
	}
	if query.SortOrder == "" {
		query.SortOrder = "desc"
	}

	// Apply sorting
	db = db.Order(query.SortBy + " " + query.SortOrder)

	// Apply filters
	for column, values := range query.Filters {
		if len(values) > 0 {
			db = db.Where(column+" IN (?)", values)
		}
	}

	// Apply pagination
	offset := (query.Page - 1) * query.PageSize
	db = db.Offset(offset).Limit(query.PageSize)

	return db
}
