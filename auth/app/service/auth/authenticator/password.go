package authenticator

import (
	"context"
	"errors"

	"github.com/Crows-Storm/Axis/auth/app/provider"
	"github.com/Crows-Storm/Axis/common/genproto/userpb"
	"github.com/Crows-Storm/Axis/common/security"
	"github.com/Crows-Storm/Axis/common/server"
)

type PasswordAuthenticator struct {
	userService      provider.UserService
	passwordVerifier PasswordVerifier
}

// PasswordVerifier nothing now
type PasswordVerifier interface {
}

func NewPasswordAuthenticator(up provider.UserService, pv PasswordVerifier) *PasswordAuthenticator {
	return &PasswordAuthenticator{userService: up, passwordVerifier: pv}
}

func (a *PasswordAuthenticator) LoginType() security.LoginType {
	return security.LoginTypePassword
}

func (a *PasswordAuthenticator) Authenticate(ctx context.Context, credential security.Credential) (*security.AuthenticatedIdentity, error) {
	c, ok := credential.(*PasswordCredential)
	if !ok {
		return nil, errors.New("invalid credential type for password authenticator")
	}

	// call gRPC to verify password, just return result(bool)
	result, err := a.userService.VerifyPassword(ctx, &userpb.VerifyPasswordRequest{
		LoginId:  c.LoginId,
		Password: c.Password,
	})
	if err != nil {
		return nil, errors.New("invalid loginId or password")
	}

	if !result.GetValue() {
		return nil, errors.New("invalid loginId or password")
	}

	// get user info to build AuthenticatedIdentity
	user, err := a.userService.GetUserByLoginId(ctx, &userpb.GetUserByLoginIdRequest{
		LoginId: c.LoginId,
	})

	if err != nil {
		return nil, errors.New(server.CodeBusinessError.String())
	}

	if user == nil {
		return nil, errors.New(server.CodeUserNotFound.String())
	}

	return &security.AuthenticatedIdentity{
		UserId:    user.Id,
		UnionId:   c.LoginId,
		Channel:   "loginId",
		IsNewUser: false,
	}, nil
}
