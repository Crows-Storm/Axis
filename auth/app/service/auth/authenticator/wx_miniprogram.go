package authenticator

// ============================================================
// WxMiniProgramAuthenticator 微信小程序登录认证器
// ============================================================

//type WxMiniProgramAuthenticator struct {
//	userService provider.UserService
//	wxClient    WxMiniProgramClient
//}
//
//// WxMiniProgramClient 微信 code2Session 调用接口（可mock测试）
//type WxMiniProgramClient interface {
//	Code2Session(ctx context.Context, code string) (*WxSessionResult, error)
//}
//
//// WxSessionResult code2Session 响应
//type WxSessionResult struct {
//	OpenID     string `json:"openid"`
//	SessionKey string `json:"session_key"`
//	UnionID    string `json:"unionid"`
//	ErrCode    int    `json:"errcode"`
//	ErrMsg     string `json:"errmsg"`
//}
//
//func NewWxMiniProgramAuthenticator(
//	up provider.UserService,
//	client WxMiniProgramClient,
//) *WxMiniProgramAuthenticator {
//	return &WxMiniProgramAuthenticator{
//		userService: up,
//		wxClient:    client,
//	}
//}
//
//func (a *WxMiniProgramAuthenticator) LoginType() security.LoginType {
//	return security.LoginTypeWxMiniProgram
//}
//
//func (a *WxMiniProgramAuthenticator) Authenticate(
//	ctx context.Context,
//	credential security.Credential,
//) (*security.AuthenticatedIdentity, error) {
//
//	c, ok := credential.(*WxMiniProgramCredential)
//	if !ok {
//		return nil, errors.New("invalid credential type for wx miniprogram authenticator")
//	}
//
//	// 1. 调用微信 code2Session
//	session, err := a.wxClient.Code2Session(ctx, c.Code)
//	if err != nil {
//		return nil, fmt.Errorf("wx code2session failed: %w", err)
//	}
//	if session.ErrCode != 0 {
//		return nil, fmt.Errorf("wx code2session error: [%d] %s", session.ErrCode, session.ErrMsg)
//	}
//	if session.OpenID == "" {
//		return nil, errors.New("wx code2session returned empty openid")
//	}
//
//	// 2. 确定统一身份标识
//	// 优先使用 unionid（跨应用统一），否则降级为 openid
//	unionID := session.UnionID
//	channel := "wx_miniprogram"
//	if unionID == "" {
//		unionID = session.OpenID
//		channel = "wx_miniprogram_openid" // 区分，后续绑定开放平台后可迁移
//	}
//
//	// 3. 查找本地用户
//	user, err := a.userService.FindByUnionID(ctx, unionID, channel)
//	isNew := false
//	if err != nil || user == nil {
//		isNew = true
//	}
//
//	identity := &security.AuthenticatedIdentity{
//		UnionId:   unionID,
//		Channel:   channel,
//		IsNewUser: isNew,
//		Extra: map[string]string{
//			"openid": session.OpenID,
//			// ⚠️ session_key 不放入 Extra，不传给前端，不存JWT
//			// 如需解密手机号，应通过单独的加密数据接口处理
//		},
//	}
//	if !isNew {
//		identity.UserId = user.UserId
//	}
//
//	return identity, nil
//}
