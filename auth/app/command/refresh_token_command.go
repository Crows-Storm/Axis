package command

// ============================================================
// RefreshTokenCommand / RefreshTokenCommandHandler
// 刷新 Token
// ============================================================

//type RefreshTokenCommand struct {
//	RefreshToken string
//}
//
//type RefreshTokenResult struct {
//	Token *jwt.TokenPayload
//}
//
//type RefreshTokenCommandHandler decorator.CommandHandler[RefreshTokenCommand, RefreshTokenResult]
//
//func NewRefreshTokenCommandHandler(
//	tokenIssuer jwt.TokenIssuer,
//) RefreshTokenCommandHandler {
//	if tokenIssuer == nil {
//		panic("nil tokenIssuer")
//	}
//
//	return decorator.ApplyCommandDecorators[RefreshTokenCommand, RefreshTokenResult](
//		refreshTokenCommandHandler{
//			tokenIssuer: tokenIssuer,
//		},
//		nil, // metricsClient: 补充后传入
//	)
//}
//
//type refreshTokenCommandHandler struct {
//	tokenIssuer jwt.TokenIssuer
//}
//
//func (h refreshTokenCommandHandler) Handle(ctx context.Context, cmd RefreshTokenCommand) (RefreshTokenResult, error) {
//	token, err := h.tokenIssuer.Refresh(ctx, cmd.RefreshToken)
//	if err != nil {
//		return RefreshTokenResult{}, fmt.Errorf("refresh token failed: %w", err)
//	}
//
//	return RefreshTokenResult{Token: token}, nil
//}
