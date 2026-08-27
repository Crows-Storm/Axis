package main

import (
	"github.com/Crows-Storm/Axis/auth/app"
	"github.com/Crows-Storm/Axis/auth/app/command"
	"github.com/Crows-Storm/Axis/auth/app/query"
	"github.com/Crows-Storm/Axis/auth/app/service/auth/authenticator"
	"github.com/Crows-Storm/Axis/common/domain/principal"
	"github.com/Crows-Storm/Axis/common/security"
	"github.com/Crows-Storm/Axis/common/server"
	"github.com/Crows-Storm/Axis/common/server/headers"
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	app app.Application
}

func (H *HTTPServer) AuthRoot(c *gin.Context) {
	// get principal from Context
	p := principal.FromContext(c.Request.Context())
	if p == nil {
		server.ErrorWithCode(c, server.CodeBadRequest)
		return
	}

	_, err := H.app.Queries.GetUser.Handle(c.Request.Context(), query.GetUserQuery{
		Id: p.UserId,
	})
	if err != nil {
		server.ErrorWithCode(c, server.CodeInvalidParams)
		return
	}

	server.Success(c, p)
	return
}

func (H *HTTPServer) Login(c *gin.Context) {
	// build req body
	var req struct {
		LoginType security.LoginType `json:"loginType" binding:"required"` // "password" | "sms_code" | "oauth" | "qrcode"
		// Use json.RawMessage to delay parsing for fields with different login methods.
		LoginId  string `json:"loginId"`
		Password string `json:"password"`
		Account  string `json:"account"`
		Code     string `json:"code"`
		Provider string `json:"provider"`
		Ticket   string `json:"ticket"`
	}
	// bind
	if err := c.ShouldBindJSON(&req); err != nil {
		server.ErrorWithCode(c, server.CodeInvalidParams)
		return
	}

	// builder Credential by login_type Strategy
	var credential security.Credential
	switch req.LoginType {
	case security.LoginTypePassword:
		credential = &authenticator.PasswordCredential{
			LoginId:  req.LoginId,
			Password: req.Password,
		}
	case security.LoginTypeSMSCode, security.LoginTypeEmailCode:
		credential = &authenticator.CodeCredential{
			LoginType_: security.LoginType(req.LoginType),
			Account:    req.Account,
			Code:       req.Code,
		}
	case security.LoginTypeOAuth:
		credential = &authenticator.OAuthCredential{
			Provider: req.Provider,
			Code:     req.Code,
		}
	case security.LoginTypeQRCode:
		credential = &authenticator.QRCodeCredential{
			Ticket: req.Ticket,
		}
	default:
		server.ErrorWithCode(c, server.CodeInvalidParams)
		return
	}

	result, err := H.app.Commands.Login.Handle(c.Request.Context(), command.LoginCommand{
		LoginType:  req.LoginType,
		Credential: credential,
	})
	if err != nil {
		server.ErrorWithCode(c, server.CodeUnauthorized)
		return
	}
	server.Success(c, result)
	return
}

func (H *HTTPServer) Logout(c *gin.Context) {
	_, err := H.app.Commands.Logout.Handle(c.Request.Context(), command.LogoutCommand{AccessToken: headers.GetAuthorization(c)})
	if err != nil {
		server.Error(c, err.Error())
		return
	}
	server.Success(c, "successfully")
	return
}

// Register is Create a New user, return create result
func (H *HTTPServer) Register(c *gin.Context) {
	var req command.RegisterUserCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		server.ErrorWithCode(c, server.CodeInternalServerError)
		return
	}
	_, err := H.app.Commands.RegisterUser.Handle(c, req)
	if err != nil {
		server.ErrorWithCodeAndMessage(c, server.CodeInternalServerError, err.Error())
		return
	}
	server.Success(c, "successfully")
	return
}
