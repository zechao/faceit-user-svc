package http_test

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zechao/faceit-user-svc/errors"
	"github.com/zechao/faceit-user-svc/http"
)

var (
	//go:embed testdata/create_request.json
	createRequest []byte
)

func TestValidateCreateUserRequest(t *testing.T) {

	t.Run("valid request", func(t *testing.T) {
		req := &http.CreateUserRequest{}
		err := json.Unmarshal(createRequest, req)
		assert.NoError(t, err)
		assert.NoError(t, req.Validate())
	})

	t.Run("invalid request", func(t *testing.T) {
		tests := []struct {
			name              string
			changeFunc        func(ss *http.CreateUserRequest)
			expectedErrDetail []errors.Detail
		}{
			{
				name: "all empty",
				changeFunc: func(r *http.CreateUserRequest) {
					r.Email = ""
					r.Password = ""
					r.FirstName = ""
					r.LastName = ""
					r.Country = ""
					r.NickName = ""
				},
				expectedErrDetail: []errors.Detail{
					{
						Field:       "first_name",
						Description: "first_name is required",
					},
					{
						Field:       "last_name",
						Description: "last_name is required",
					},
					{
						Field:       "nick_name",
						Description: "nick_name is required",
					},

					{
						Field:       "password",
						Description: "password is required",
					},
					{
						Field:       "email",
						Description: "email is required",
					},
					{
						Field:       "country",
						Description: "country must be a 2-letter ISO country code",
					},
				},
			},
			{
				name: "invalid password length",
				changeFunc: func(r *http.CreateUserRequest) {
					r.Password = "123456"

				},
				expectedErrDetail: []errors.Detail{
					{
						Field:       "password",
						Description: "password must between 8 and 40 characters long",
					},
				},
			},
			{
				name: "invalid email format",
				changeFunc: func(r *http.CreateUserRequest) {
					r.Email = "adas.com"

				},
				expectedErrDetail: []errors.Detail{
					{
						Field:       "email",
						Description: "invalid email format",
					},
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var req http.CreateUserRequest
				err := json.Unmarshal(createRequest, &req)
				assert.NoError(t, err)
				tt.changeFunc(&req)
				err = req.Validate()

				wrongInputErr := new(errors.Error)
				errors.As(err, &wrongInputErr)
				assert.ElementsMatch(t, tt.expectedErrDetail, wrongInputErr.Details)
			})
		}
	})

}



