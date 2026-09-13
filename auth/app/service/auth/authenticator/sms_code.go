package authenticator

//type SMSCodeAuthenticator struct {
//	userService  provider.UserService
//	codeVerifier CodeVerifier
//}
//
//type CodeVerifier interface {
//	Verify(ctx context.Context, account string, code string, purpose string) (bool, error)
//}
//
//func NewSMSCodeAuthenticator(up provider.UserService, cv CodeVerifier) *SMSCodeAuthenticator {
//	return &SMSCodeAuthenticator{userService: up, codeVerifier: cv}
//}
//
//func (a *SMSCodeAuthenticator) LoginType() security.LoginType {
//	return security.LoginTypeSMSCode
//}
//
//func (a *SMSCodeAuthenticator) Authenticate(ctx context.Context, credential security.Credential) (*security.AuthenticatedIdentity, error) {
//	c, ok := credential.(*CodeCredential)
//	if !ok {
//		return nil, errors.New("invalid credential type for SMS authenticator")
//	}
//
//	valid, err := a.codeVerifier.Verify(ctx, c.Account, c.Code, "login")
//	if err != nil || !valid {
//		return nil, errors.New("invalid or expired verification code")
//	}
//
//	user, err := a.userService.FindByUnionID(ctx, c.Account, "phone")
//	isNew := false
//	if err != nil || user == nil {
//		isNew = true
//	}
//
//	identity := &security.AuthenticatedIdentity{
//		UnionId:   c.Account,
//		Channel:   "phone",
//		IsNewUser: isNew,
//	}
//	if !isNew {
//		identity.UserId = user.UserId
//	}
//
//	return identity, nil
//}
