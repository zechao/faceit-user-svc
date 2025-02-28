package query_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zechao/faceit-user-svc/errors"
	"github.com/zechao/faceit-user-svc/query"
)

func TestQueryFromURLFail(t *testing.T) {
	tableTest := map[string]struct {
		inputQuery    string
		expectedError error
	}{
		"invalid page value": {
			inputQuery: "page=abc",
			expectedError: errors.NewWrongInput(query.ErrCodeInvalidParameter, errors.Detail{
				Field:       "page",
				Description: "page must be a number",
			}),
		},
		"invalid page number": {
			inputQuery: "page=0",
			expectedError: errors.NewWrongInput(query.ErrCodeInvalidParameter, errors.Detail{
				Field:       "page",
				Description: "page number must be greater than 0",
			}),
		},
		"invalid page_size": {
			inputQuery: "page=1&page_size=abc",
			expectedError: errors.NewWrongInput(query.ErrCodeInvalidParameter, errors.Detail{
				Field:       "page_size",
				Description: "page_size must be a number",
			}),
		},
		"invalid page_size value": {
			inputQuery: "page=1&page_size=abc",
			expectedError: errors.NewWrongInput(query.ErrCodeInvalidParameter, errors.Detail{
				Field:       "page_size",
				Description: "page_size must be a number",
			}),
		},
		"invalid page_size number": {
			inputQuery: "page=1&page_size=0",
			expectedError: errors.NewWrongInput(query.ErrCodeInvalidParameter, errors.Detail{
				Field:       "page_size",
				Description: "page_size must be greater than 0",
			}),
		},
		"invalid sort_order value": {
			inputQuery: "page=1&page_size=11&sort_order=invalid",
			expectedError: errors.NewWrongInput(query.ErrCodeInvalidParameter, errors.Detail{
				Field:       "sort_order",
				Description: "sort_order must be desc or asc",
			}),
		},
		"unsupported filter": {
			inputQuery: "a=b",
			expectedError: errors.NewWrongInput(query.ErrCodeInvalidParameter, errors.Detail{
				Field:       "a",
				Description: "parameter a is not supported",
			}),
		},
	}
	for name, testCase := range tableTest {
		t.Run(name, func(t *testing.T) {
			request, err := http.NewRequest("GET", "?"+testCase.inputQuery, nil)
			assert.NoError(t, err)
			q, err := query.QueryFromURL(request.URL.Query())
			assert.Nil(t, q)
			assert.Equal(t, testCase.expectedError, err)
		})
	}
}

func TestQueryFromURLSuccess(t *testing.T) {
	tableTest := map[string]struct {
		inputQuery    string
		expectedQuery *query.Query
	}{
		"default params": {
			inputQuery: "page=1",
			expectedQuery: &query.Query{
				Page:      1,
				PageSize:  100,
				SortOrder: "desc",
				SortBy:    "created_at",
				Filters:   make(map[string][]string),
			},
		},
		"all set": {
			inputQuery: "page=10&page_size=10&sort_order=asc&sort_by=name&country=UK&country=ES&first_name=zechao",
			expectedQuery: &query.Query{
				Page:      10,
				PageSize:  10,
				SortOrder: "asc",
				SortBy:    "name",
				Filters: map[string][]string{
					"country":    {"UK", "ES"},
					"first_name": {"zechao"},
				},
			},
		},
	}
	for name, testCase := range tableTest {
		t.Run(name, func(t *testing.T) {
			request, err := http.NewRequest("GET", "?"+testCase.inputQuery, nil)
			assert.NoError(t, err)
			q, err := query.QueryFromURL(request.URL.Query())
			assert.Nil(t, err)
			assert.Equal(t, q, testCase.expectedQuery)
		})
	}
}
