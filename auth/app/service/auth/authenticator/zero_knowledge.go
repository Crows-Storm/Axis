package authenticator

import (
	"context"

	"github.com/Crows-Storm/Axis/auth/app/provider"
	"github.com/Crows-Storm/Axis/common/security"
)

// The ZeroKnowledgeAuthenticator is use zero-knowledge to authenticator
// TODO
type ZeroKnowledgeAuthenticator struct {
	userService provider.UserService
}

func NewZeroKnowledgeAuthenticator(up provider.UserService, pv PasswordVerifier) *ZeroKnowledgeAuthenticator {
	return &ZeroKnowledgeAuthenticator{userService: up}
}

func (a *ZeroKnowledgeAuthenticator) LoginType() security.LoginType {
	return security.LoginTypePassword
}

func (a *ZeroKnowledgeAuthenticator) Authenticate(ctx context.Context, credential security.Credential) (*security.AuthenticatedIdentity, error) {
	return &security.AuthenticatedIdentity{
		UserId:    0,
		UnionId:   "",
		Channel:   "ZK",
		IsNewUser: false,
	}, nil
}
