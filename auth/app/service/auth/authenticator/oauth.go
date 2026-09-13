package authenticator

//type OAuthProvider interface {
//	// ExchangeToken 用 authorization code 换取 access_token
//	ExchangeToken(ctx context.Context, code string) (*OAuthTokenResult, error)
//	// GetUserInfo 用 access_token 获取用户信息
//	GetUserInfo(ctx context.Context, accessToken string) (*OAuthUserInfo, error)
//	// ProviderName 提供商标识
//	ProviderName() string
//}
//
//type OAuthTokenResult struct {
//	AccessToken  string
//	RefreshToken string
//	ExpiresIn    int64
//}
//
//type OAuthUserInfo struct {
//	UnionID  string
//	Nickname string
//	Avatar   string
//	Email    string
//	Phone    string
//}
//
//type OAuthAuthenticator struct {
//	userService provider.UserService
//	providers   map[string]OAuthProvider // "wechat" -> WechatProvider, "github" -> GithubProvider
//}
//
//func NewOAuthAuthenticator(up provider.UserService, providers ...OAuthProvider) *OAuthAuthenticator {
//	m := make(map[string]OAuthProvider, len(providers))
//	for _, p := range providers {
//		m[p.ProviderName()] = p
//	}
//	return &OAuthAuthenticator{userService: up, providers: m}
//}
//
//func (a *OAuthAuthenticator) LoginType() security.LoginType {
//	return security.LoginTypeOAuth
//}
//
//func (a *OAuthAuthenticator) Authenticate(ctx context.Context, credential security.Credential) (*security.AuthenticatedIdentity, error) {
//	c, ok := credential.(*OAuthCredential)
//	if !ok {
//		return nil, errors.New("invalid credential type for OAuth authenticator")
//	}
//
//	// 1. 获取对应的 OAuth Provider
//	provider, exists := a.providers[c.Provider]
//	if !exists {
//		return nil, fmt.Errorf("unsupported oauth provider: %s", c.Provider)
//	}
//
//	// 2. Code 换 Token
//	tokenResult, err := provider.ExchangeToken(ctx, c.Code)
//	if err != nil {
//		return nil, fmt.Errorf("oauth exchange token failed: %w", err)
//	}
//
//	// 3. Token 换用户信息
//	oauthUser, err := provider.GetUserInfo(ctx, tokenResult.AccessToken)
//	if err != nil {
//		return nil, fmt.Errorf("oauth get user info failed: %w", err)
//	}
//
//	user, err := a.userService.GetUserById(ctx, oauthUser.UnionID, c.Provider)
//	isNew := false
//	if err != nil || user == nil {
//		isNew = true
//	}
//
//	identity := &security.AuthenticatedIdentity{
//		UnionId:   oauthUser.UnionID,
//		Channel:   c.Provider,
//		IsNewUser: isNew,
//		Extra: map[string]string{
//			"nickname": oauthUser.Nickname,
//			"avatar":   oauthUser.Avatar,
//			"email":    oauthUser.Email,
//		},
//	}
//	if !isNew {
//		identity.UserId = user.UserId
//	}
//
//	return identity, nil
//}
