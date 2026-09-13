package authenticator

// QRCodeRepository QR code status storage interface (usually based on Redis)
//
// QR code login process:
// 1. Web client: Generate ticket → Display QR code → Poll ticket status
// 2. App client: Scan QR code → Confirm authorization → Write userID to ticket
// 3. Web client: Poll to find authorized → Obtain userID → Issue Token
//type QRCodeRepository interface {
//	// CreateTicket 创建待扫码的 ticket
//	CreateTicket(ctx context.Context, ttl time.Duration) (ticket string, err error)
//	// Confirm 扫码确认（App端调用），绑定 userID
//	Confirm(ctx context.Context, ticket string, userID int64) error
//	// GetConfirmedUserId 检查 ticket 是否已被确认，返回 userID
//	GetConfirmedUserId(ctx context.Context, ticket string) (userID int64, confirmed bool, err error)
//	// Expire 使 ticket 失效
//	Expire(ctx context.Context, ticket string) error
//}
//
//type QRCodeAuthenticator struct {
//	userService provider.UserService
//	qrCodeStore QRCodeRepository
//}
//
//func NewQRCodeAuthenticator(up provider.UserService, store QRCodeRepository) *QRCodeAuthenticator {
//	return &QRCodeAuthenticator{userService: up, qrCodeStore: store}
//}
//
//func (a *QRCodeAuthenticator) LoginType() security.LoginType {
//	return security.LoginTypeQRCode
//}
//
//func (a *QRCodeAuthenticator) Authenticate(ctx context.Context, credential security.Credential) (*security.AuthenticatedIdentity, error) {
//	c, ok := credential.(*QRCodeCredential)
//	if !ok {
//		return nil, errors.New("invalid credential type for QRCode authenticator")
//	}
//
//	// 1. 检查 ticket 是否已被扫码确认
//	userID, confirmed, err := a.qrCodeStore.GetConfirmedUserId(ctx, c.Ticket)
//	if err != nil {
//		return nil, errors.New("qrcode ticket check failed")
//	}
//	if !confirmed {
//		return nil, ErrQRCodeNotConfirmed // Web端继续轮询
//	}
//
//	// 2. 使 ticket 失效（一次性使用）
//	_ = a.qrCodeStore.Expire(ctx, c.Ticket)
//
//	// 3. 查找用户
//	user, err := a.userService.FindByID(ctx, userID)
//	if err != nil {
//		return nil, errors.New("user not found for qrcode login")
//	}
//
//	return &security.AuthenticatedIdentity{
//		UserId:    user.UserId,
//		UnionId:   user.Username,
//		Channel:   "qrcode",
//		IsNewUser: false, // 扫码登录不存在新用户
//	}, nil
//}
//
//// ---- 扫码登录专用辅助方法 ----
//
//// GenerateTicket 生成二维码 ticket（供 Handler 调用）
//func (a *QRCodeAuthenticator) GenerateTicket(ctx context.Context) (string, error) {
//	return a.qrCodeStore.CreateTicket(ctx, 5*time.Minute) // 5分钟过期
//}
//
//var ErrQRCodeNotConfirmed = errors.New("qrcode not yet confirmed, please retry")
