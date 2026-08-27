package auth

import (
	"github.com/Crows-Storm/Axis/auth/app/provider"
	"github.com/Crows-Storm/Axis/auth/app/service/auth/authenticator"
	"github.com/Crows-Storm/Axis/common/jwt"
)

// InitAuthAppService init the auth strategy service
func InitAuthAppService(
	userService provider.UserService,
	jwtIssuer *jwt.JWTIssuer,
	// codeVerifier authenticator.CodeVerifier,
	// qrCodeStore authenticator.QRCodeRepository,
) *AuthAppService {

	// TODO: Provider need implement OAuthProvider or PasswordVerifier interface to verify in the redis repository
	passwordAuth := authenticator.NewPasswordAuthenticator(userService, nil) // passwordVerifier

	//smsAuth := authenticator.NewSMSCodeAuthenticator(userService, codeVerifier)
	//oauthAuth := authenticator.NewOAuthAuthenticator(userService,
	//	nil, // wechatProvider
	//	nil, // githubProvider
	//)
	//qrCodeAuth := authenticator.NewQRCodeAuthenticator(userService, qrCodeStore)
	//wxMiniprogram := authenticator.NewWxMiniProgramAuthenticator(userService, nil) // wxMiniprogramProvider

	return NewAuthAppService(
		jwtIssuer,
		passwordAuth,
		//smsAuth,
		//oauthAuth,
		//qrCodeAuth,
		//wxMiniprogram,
	)
}
