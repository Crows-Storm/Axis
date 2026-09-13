package command

// ============================================================
// RegisterAndLoginCommand / RegisterAndLoginCommandHandler
// 注册并登录（验证码登录 / OAuth 首次登录场景）
//
// 流程: 认证 → 发现新用户 → 自动注册 → 加载权限 → 签发Token
// ============================================================

//type RegisterAndLoginCommand struct {
//	LoginType  security.LoginType
//	Credential security.Credential
//	Username   string `json:"username"`
//	Nickname   string `json:"nickname"`
//}
//
//type RegisterAndLoginResult struct {
//	UserId   int64             `json:"user_id"`
//	Username string            `json:"username"`
//	IsNew    bool              `json:"is_new"`
//	Token    *jwt.TokenPayload `json:"token"`
//}
//
//type RegisterAndLoginCommandHandler decorator.CommandHandler[RegisterAndLoginCommand, RegisterAndLoginResult]
//
//func NewRegisterAndLoginCommandHandler(
//	userService provider.UserService,
//	tokenIssuer jwt.TokenIssuer,
//	authenticators ...security.Authenticator,
//) RegisterAndLoginCommandHandler {
//	if userService == nil {
//		panic("nil userService")
//	}
//	if tokenIssuer == nil {
//		panic("nil tokenIssuer")
//	}
//
//	m := make(map[security.LoginType]security.Authenticator, len(authenticators))
//	for _, a := range authenticators {
//		m[a.LoginType()] = a
//	}
//
//	return decorator.ApplyCommandDecorators[RegisterAndLoginCommand, RegisterAndLoginResult](
//		registerAndLoginCommandHandler{
//			userService:    userService,
//			tokenIssuer:    tokenIssuer,
//			authenticators: m,
//		},
//		nil, // metricsClient: 补充后传入
//	)
//}
//
//type registerAndLoginCommandHandler struct {
//	userService    provider.UserService
//	tokenIssuer    jwt.TokenIssuer
//	authenticators map[security.LoginType]security.Authenticator
//}
//
//func (h registerAndLoginCommandHandler) Handle(ctx context.Context, cmd RegisterAndLoginCommand) (RegisterAndLoginResult, error) {
//
//	// 1. 选择认证策略
//	a, ok := h.authenticators[cmd.LoginType]
//	if !ok {
//		return RegisterAndLoginResult{}, fmt.Errorf("unsupported login type: %s", cmd.LoginType)
//	}
//
//	// 2. 凭证校验
//	if err := cmd.Credential.Validate(); err != nil {
//		return RegisterAndLoginResult{}, fmt.Errorf("credential validation failed: %w", err)
//	}
//
//	// 3. 执行认证
//	identity, err := a.Authenticate(ctx, cmd.Credential)
//	if err != nil {
//		return RegisterAndLoginResult{}, fmt.Errorf("authentication failed: %w", err)
//	}
//
//	// 4. 如果不是新用户，直接登录
//	if !identity.IsNewUser {
//		principal, err := h.userService.LoadPrincipal(ctx, identity.UserId, identity.UnionId)
//		if err != nil {
//			return RegisterAndLoginResult{}, fmt.Errorf("load principal failed: %w", err)
//		}
//
//		token, err := h.tokenIssuer.Issue(ctx, principal)
//		if err != nil {
//			return RegisterAndLoginResult{}, fmt.Errorf("issue token failed: %w", err)
//		}
//
//		return RegisterAndLoginResult{
//			UserId:   identity.UserId,
//			Username: principal.Username,
//			IsNew:    false,
//			Token:    token,
//		}, nil
//	}
//
//	// 5. 注册新用户
//	nickname := cmd.Nickname
//	if nickname == "" && identity.Extra != nil {
//		nickname = identity.Extra["nickname"] // OAuth 场景自动填充
//	}
//
//	userResp, err := h.userService.CreateAndBindIdentity(ctx, &auth.CreateUserCommand{
//		Username: cmd.Username,
//		Nickname: nickname,
//		UnionID:  identity.UnionId,
//		Channel:  identity.Channel,
//		Extra:    identity.Extra,
//	})
//	if err != nil {
//		return RegisterAndLoginResult{}, fmt.Errorf("register user failed: %w", err)
//	}
//
//	// 6. 加载安全上下文
//	principal, err := h.userService.LoadPrincipal(ctx, userResp.UserId, userResp.Username)
//	if err != nil {
//		return RegisterAndLoginResult{}, fmt.Errorf("load principal failed: %w", err)
//	}
//
//	// 7. 签发 Token
//	token, err := h.tokenIssuer.Issue(ctx, principal)
//	if err != nil {
//		return RegisterAndLoginResult{}, fmt.Errorf("issue token failed: %w", err)
//	}
//
//	return RegisterAndLoginResult{
//		UserId:   userResp.UserId,
//		Username: userResp.Username,
//		IsNew:    true,
//		Token:    token,
//	}, nil
//}
//
//// localPrincipal is a convenience type alias used within this handler
//// to avoid import cycles where principal is already imported.
//type localPrincipal = principal.Principal
