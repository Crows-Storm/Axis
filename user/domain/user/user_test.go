package user

import (
	"errors"
	"testing"

	commuser "github.com/Crows-Storm/Axis/common/domain/user"
	"github.com/stretchr/testify/assert"
)

func TestUser_Create(t *testing.T) {
	tests := []struct {
		name string

		loginID  string
		email    string
		username string

		wantStatus  commuser.Status
		wantVersion int64

		wantErrContain string
		wantErr        error
	}{
		{
			name: "Create User Success",

			loginID:  "test1",
			email:    "dadas@163.com",
			username: "test1Name",

			wantStatus:  commuser.Activated,
			wantVersion: 1,

			wantErrContain: "",
			wantErr:        nil,
		},
		{
			name: "Create User Success",

			loginID:  "test1",
			email:    "dadas@163.com",
			username: "test1Name",

			wantStatus:     commuser.Activated,
			wantVersion:    1,
			wantErrContain: "user.Create: failed to apply domain event",
			wantErr:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			u, err := Create(tt.loginID, tt.email, tt.username)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr != nil {
				return
			}

			// assert
			assert.Equal(t, tt.loginID, u.LoginID)
			assert.Equal(t, tt.email, u.Email)
			assert.Equal(t, tt.username, u.Username)
			assert.Equal(t, tt.wantStatus, u.Status)
			assert.Equal(t, tt.wantVersion, u.Version())
			assert.Equal(t, 1, len(u.Events()))

			assert.Nil(t, err)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
