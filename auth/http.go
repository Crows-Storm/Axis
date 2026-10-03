package main

import (
	"github.com/Crows-Storm/Axis/auth/app"
	"github.com/Crows-Storm/Axis/auth/app/command"
	"github.com/Crows-Storm/Axis/auth/app/query"
	"github.com/Crows-Storm/Axis/auth/app/service/auth/authenticator"
	"github.com/Crows-Storm/Axis/auth/domain/menu"
	"github.com/Crows-Storm/Axis/auth/domain/role"
	"github.com/Crows-Storm/Axis/auth/protos"
	"github.com/Crows-Storm/Axis/common/contextx"
	"github.com/Crows-Storm/Axis/common/domain/principal"
	"github.com/Crows-Storm/Axis/common/security"
	"github.com/Crows-Storm/Axis/common/server"
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	app app.Application
}

var _ protos.ServerInterface = (*HTTPServer)(nil)

func (H *HTTPServer) AuthRoot(c *gin.Context) {
	// get principal from Context
	p := principal.FromContext(c.Request.Context())
	if p == nil {
		server.ErrorWithCode(c, server.CodeBadRequest)
		return
	}

	_, err := H.app.Queries.GetUser.Handle(c.Request.Context(), query.GetUserQuery{
		ID: p.UserID,
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
		LoginType security.LoginType `json:"login_type" binding:"required"` // "password" | "sms_code" | "oauth" | "qrcode"
		// Use json.RawMessage to delay parsing for fields with different login methods.
		LoginID  string `json:"login_id"`
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
			LoginID:  req.LoginID,
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
	_, err := H.app.Commands.Logout.Handle(c.Request.Context(), command.LogoutCommand{AccessToken: contextx.FromGinContext(c).Authorization()})
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
	_, err := H.app.Commands.RegisterUser.Handle(c.Request.Context(), req)
	if err != nil {
		server.ErrorWithCodeAndMessage(c, server.CodeInternalServerError, err.Error())
		return
	}
	server.Success(c, "successfully")
}

func currentUserID(c *gin.Context) (uint64, bool) {
	p := principal.FromContext(c.Request.Context())
	if p == nil || p.UserID == 0 {
		return 0, false
	}
	return p.UserID, true
}

func (H *HTTPServer) GetUserMenus(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		server.ErrorWithCode(c, server.CodeInvalidToken)
		return
	}
	tree, err := H.app.Queries.GetUserMenuTree.Handle(c.Request.Context(), query.GetUserMenuTreeQuery{UserID: userID})
	if err != nil {
		server.ErrorWithCodeAndMessage(c, server.CodeInternalServerError, err.Error())
		return
	}
	server.Success(c, tree)
}

func (H *HTTPServer) GetUserRoles(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		server.ErrorWithCode(c, server.CodeInvalidToken)
		return
	}
	roles, err := H.app.Queries.GetUserRoles.Handle(c.Request.Context(), query.GetUserRolesQuery{UserID: userID})
	if err != nil {
		server.ErrorWithCodeAndMessage(c, server.CodeInternalServerError, err.Error())
		return
	}
	server.Success(c, roles)
}

func (H *HTTPServer) CheckPermission(c *gin.Context, params protos.CheckPermissionParams) {
	userID, ok := currentUserID(c)
	if !ok {
		server.ErrorWithCode(c, server.CodeInvalidToken)
		return
	}
	result, err := H.app.Queries.CheckMenuPermission.Handle(c.Request.Context(), query.CheckMenuPermissionQuery{
		UserID:   userID,
		MenuCode: params.Menu,
		Action:   string(params.Action),
	})
	if err != nil {
		server.ErrorWithCodeAndMessage(c, server.CodeBadRequest, err.Error())
		return
	}
	server.Success(c, result)
}

func (H *HTTPServer) AssignUserRole(c *gin.Context) {
	var req command.AssignUserRoleCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		server.ErrorWithCode(c, server.CodeInvalidParams)
		return
	}
	if _, err := H.app.Commands.AssignUserRole.Handle(c.Request.Context(), req); err != nil {
		server.ErrorWithCodeAndMessage(c, server.CodeBadRequest, err.Error())
		return
	}
	server.Success(c, "successfully")
}

func (H *HTTPServer) UpdateRoleMenuMask(c *gin.Context) {
	var req command.UpdateRoleMenuMaskCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		server.ErrorWithCode(c, server.CodeInvalidParams)
		return
	}
	if _, err := H.app.Commands.UpdateRoleMenuMask.Handle(c.Request.Context(), req); err != nil {
		server.ErrorWithCodeAndMessage(c, server.CodeBadRequest, err.Error())
		return
	}
	server.Success(c, "successfully")
}

func pathID(id uint64) (uint64, bool) {
	if id == 0 {
		return 0, false
	}
	return id, true
}

func toRoleView(r *role.Role) query.RoleView {
	return query.RoleView{
		ID:          r.ID,
		Code:        r.Code,
		Name:        r.Name,
		Description: r.Description,
		Status:      int8(r.Status),
	}
}

func toMenuNode(m *menu.Menu) *menu.Node {
	return m.ToNode(0)
}

func (H *HTTPServer) UnassignUserRole(c *gin.Context) {
	var req command.UnassignUserRoleCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		server.ErrorWithCode(c, server.CodeInvalidParams)
		return
	}
	if _, err := H.app.Commands.UnassignUserRole.Handle(c.Request.Context(), req); err != nil {
		server.ErrorWithCodeAndMessage(c, server.CodeBadRequest, err.Error())
		return
	}
	server.Success(c, "successfully")
}

func (H *HTTPServer) ReplaceUserRoles(c *gin.Context) {
	var req command.ReplaceUserRolesCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		server.ErrorWithCode(c, server.CodeInvalidParams)
		return
	}
	if _, err := H.app.Commands.ReplaceUserRoles.Handle(c.Request.Context(), req); err != nil {
		server.ErrorWithCodeAndMessage(c, server.CodeBadRequest, err.Error())
		return
	}
	server.Success(c, "successfully")
}

func (H *HTTPServer) RevokeRoleMenu(c *gin.Context) {
	var req command.RevokeRoleMenuCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		server.ErrorWithCode(c, server.CodeInvalidParams)
		return
	}
	if _, err := H.app.Commands.RevokeRoleMenu.Handle(c.Request.Context(), req); err != nil {
		server.ErrorWithCodeAndMessage(c, server.CodeBadRequest, err.Error())
		return
	}
	server.Success(c, "successfully")
}

func (H *HTTPServer) GetAssignedUserRoles(c *gin.Context, userId uint64) {
	userID, ok := pathID(userId)
	if !ok {
		server.ErrorWithCode(c, server.CodeInvalidParams)
		return
	}
	roles, err := H.app.Queries.ListAssignedRoles.Handle(c.Request.Context(), query.ListAssignedRolesQuery{UserID: userID})
	if err != nil {
		server.ErrorWithCodeAndMessage(c, server.CodeInternalServerError, err.Error())
		return
	}
	server.Success(c, roles)
}

func (H *HTTPServer) ListRoles(c *gin.Context) {
	roles, err := H.app.Queries.ListRoles.Handle(c.Request.Context(), query.ListRolesQuery{})
	if err != nil {
		server.ErrorWithCodeAndMessage(c, server.CodeInternalServerError, err.Error())
		return
	}
	server.Success(c, roles)
}

func (H *HTTPServer) CreateRole(c *gin.Context) {
	var req command.CreateRoleCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		server.ErrorWithCode(c, server.CodeInvalidParams)
		return
	}
	created, err := H.app.Commands.CreateRole.Handle(c.Request.Context(), req)
	if err != nil {
		server.ErrorWithCodeAndMessage(c, server.CodeBadRequest, err.Error())
		return
	}
	server.Success(c, toRoleView(created))
}

func (H *HTTPServer) GetRole(c *gin.Context, roleId uint64) {
	id, ok := pathID(roleId)
	if !ok {
		server.ErrorWithCode(c, server.CodeInvalidParams)
		return
	}
	detail, err := H.app.Queries.GetRole.Handle(c.Request.Context(), query.GetRoleQuery{RoleID: id})
	if err != nil {
		server.ErrorWithCodeAndMessage(c, server.CodeNotFound, err.Error())
		return
	}
	server.Success(c, detail)
}

func (H *HTTPServer) UpdateRole(c *gin.Context, roleId uint64) {
	id, ok := pathID(roleId)
	if !ok {
		server.ErrorWithCode(c, server.CodeInvalidParams)
		return
	}
	var req command.UpdateRoleCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		server.ErrorWithCode(c, server.CodeInvalidParams)
		return
	}
	req.RoleID = id
	if _, err := H.app.Commands.UpdateRole.Handle(c.Request.Context(), req); err != nil {
		server.ErrorWithCodeAndMessage(c, server.CodeBadRequest, err.Error())
		return
	}
	detail, err := H.app.Queries.GetRole.Handle(c.Request.Context(), query.GetRoleQuery{RoleID: id})
	if err != nil {
		server.ErrorWithCodeAndMessage(c, server.CodeNotFound, err.Error())
		return
	}
	server.Success(c, detail.RoleView)
}

func (H *HTTPServer) ChangeRoleStatus(c *gin.Context, roleId uint64) {
	id, ok := pathID(roleId)
	if !ok {
		server.ErrorWithCode(c, server.CodeInvalidParams)
		return
	}
	var req command.UpdateRoleStatusCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		server.ErrorWithCode(c, server.CodeInvalidParams)
		return
	}
	req.RoleID = id
	if _, err := H.app.Commands.UpdateRoleStatus.Handle(c.Request.Context(), req); err != nil {
		server.ErrorWithCodeAndMessage(c, server.CodeBadRequest, err.Error())
		return
	}
	detail, err := H.app.Queries.GetRole.Handle(c.Request.Context(), query.GetRoleQuery{RoleID: id})
	if err != nil {
		server.ErrorWithCodeAndMessage(c, server.CodeNotFound, err.Error())
		return
	}
	server.Success(c, detail.RoleView)
}

func (H *HTTPServer) GetRoleMenus(c *gin.Context, roleId uint64) {
	id, ok := pathID(roleId)
	if !ok {
		server.ErrorWithCode(c, server.CodeInvalidParams)
		return
	}
	detail, err := H.app.Queries.GetRole.Handle(c.Request.Context(), query.GetRoleQuery{RoleID: id})
	if err != nil {
		server.ErrorWithCodeAndMessage(c, server.CodeNotFound, err.Error())
		return
	}
	server.Success(c, detail.Menus)
}

func (H *HTTPServer) ListMenus(c *gin.Context) {
	tree, err := H.app.Queries.ListMenuCatalog.Handle(c.Request.Context(), query.ListMenuCatalogQuery{})
	if err != nil {
		server.ErrorWithCodeAndMessage(c, server.CodeInternalServerError, err.Error())
		return
	}
	server.Success(c, tree)
}

func (H *HTTPServer) CreateMenu(c *gin.Context) {
	var req command.CreateMenuCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		server.ErrorWithCode(c, server.CodeInvalidParams)
		return
	}
	created, err := H.app.Commands.CreateMenu.Handle(c.Request.Context(), req)
	if err != nil {
		server.ErrorWithCodeAndMessage(c, server.CodeBadRequest, err.Error())
		return
	}
	server.Success(c, toMenuNode(created))
}

func (H *HTTPServer) UpdateMenu(c *gin.Context, menuId uint64) {
	id, ok := pathID(menuId)
	if !ok {
		server.ErrorWithCode(c, server.CodeInvalidParams)
		return
	}
	var req command.UpdateMenuCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		server.ErrorWithCode(c, server.CodeInvalidParams)
		return
	}
	req.MenuID = id
	if _, err := H.app.Commands.UpdateMenu.Handle(c.Request.Context(), req); err != nil {
		server.ErrorWithCodeAndMessage(c, server.CodeBadRequest, err.Error())
		return
	}
	current, err := H.app.Queries.GetMenu.Handle(c.Request.Context(), query.GetMenuQuery{MenuID: id})
	if err != nil {
		server.ErrorWithCodeAndMessage(c, server.CodeNotFound, err.Error())
		return
	}
	server.Success(c, toMenuNode(current))
}

func (H *HTTPServer) ChangeMenuStatus(c *gin.Context, menuId uint64) {
	id, ok := pathID(menuId)
	if !ok {
		server.ErrorWithCode(c, server.CodeInvalidParams)
		return
	}
	var req command.UpdateMenuStatusCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		server.ErrorWithCode(c, server.CodeInvalidParams)
		return
	}
	req.MenuID = id
	if _, err := H.app.Commands.UpdateMenuStatus.Handle(c.Request.Context(), req); err != nil {
		server.ErrorWithCodeAndMessage(c, server.CodeBadRequest, err.Error())
		return
	}
	current, err := H.app.Queries.GetMenu.Handle(c.Request.Context(), query.GetMenuQuery{MenuID: id})
	if err != nil {
		server.ErrorWithCodeAndMessage(c, server.CodeNotFound, err.Error())
		return
	}
	server.Success(c, toMenuNode(current))
}

func (H *HTTPServer) ListActions(c *gin.Context) {
	actions, err := H.app.Queries.ListActions.Handle(c.Request.Context(), query.ListActionsQuery{})
	if err != nil {
		server.ErrorWithCodeAndMessage(c, server.CodeInternalServerError, err.Error())
		return
	}
	server.Success(c, actions)
}
