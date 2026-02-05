package api

import (
	"bytes"
	"encoding/json"
	"time"

	"bbs-go/internal/controllers/render"
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/config"
	"bbs-go/internal/services"

	"github.com/coreos/go-oidc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web"
	"golang.org/x/oauth2"
)

type OIDCController struct {
	Ctx iris.Context
}

// GetSignin 发起 OIDC 登录
func (c *OIDCController) GetSignin() *web.JsonResult {
	conf := config.Instance.OIDC
	if conf.ClientID == "" || conf.Issuer == "" {
		return web.JsonErrorMsg("OIDC not configured")
	}
	oidcCtx := c.Ctx.Request().Context()
	provider, err := oidc.NewProvider(oidcCtx, conf.Issuer)
	if err != nil {
		return web.JsonError(err)
	}
	oauth2Config := oauth2.Config{
		ClientID:     conf.ClientID,
		ClientSecret: conf.SecretKey,
		RedirectURL:  conf.Callback,
		Endpoint:     provider.Endpoint(),
		Scopes: []string{
			oidc.ScopeOpenID,
			"profile",
			"email",
			"groups",
			"offline_access",
			"user_id",
			"employee_no",
			"login_type",
		},
	}
	state := strs.UUID()
	redirect := c.Ctx.FormValue("redirect")
	// 使用 JWT 生成 state token，包含 state 和 redirect 信息，并设置过期时间
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"state":    state,
		"redirect": redirect,
		"exp":      time.Now().Add(10 * time.Minute).Unix(),
	})
	stateToken, err := token.SignedString([]byte(conf.SecretKey))
	if err != nil {
		return web.JsonError(err)
	}
	// 将 state token 存入 Cookie
	c.Ctx.SetCookieKV(
		"oidc_state_token",
		stateToken,
		iris.CookieExpires(10*time.Minute),
		iris.CookiePath("/"),
	)
	url := oauth2Config.AuthCodeURL(state)
	if c.Ctx.FormValue("type") != "" {
		url += "&type=" + c.Ctx.FormValue("type")
	}
	c.Ctx.Redirect(url)
	return nil
}

// GetCallback OIDC 登录回调
func (c *OIDCController) GetCallback() *web.JsonResult {
	conf := config.Instance.OIDC
	incomingState := c.Ctx.URLParam("state")
	if incomingState == "" {
		return web.JsonErrorMsg("OIDC state not found")
	}
	stateToken := c.Ctx.GetCookie("oidc_state_token")
	if stateToken == "" {
		return web.JsonErrorMsg("OIDC state token not found")
	}
	// 验证 State Token
	token, err := jwt.Parse(stateToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(conf.SecretKey), nil
	})
	if err != nil || !token.Valid {
		return web.JsonErrorMsg("Invalid OIDC state token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return web.JsonErrorMsg("Invalid OIDC state token claims")
	}
	expectedState := claims["state"].(string)
	if expectedState != incomingState {
		return web.JsonErrorMsg("OIDC state mismatch")
	}
	ctx := c.Ctx.Request().Context()
	provider, err := oidc.NewProvider(ctx, conf.Issuer)
	if err != nil {
		return web.JsonError(err)
	}
	oauth2Config := oauth2.Config{
		ClientID:     conf.ClientID,
		ClientSecret: conf.SecretKey,
		RedirectURL:  conf.Callback,
		Endpoint:     provider.Endpoint(),
		Scopes: []string{
			oidc.ScopeOpenID,
			"profile",
			"email",
			"groups",
			"offline_access",
			"user_id",
			"employee_no",
			"login_type"},
	}
	// 使用 Code 换取 Token
	oauth2Token, err := oauth2Config.Exchange(ctx, c.Ctx.URLParam("code"))
	if err != nil {
		return web.JsonError(err)
	}
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return web.JsonErrorMsg("No id_token field in oauth2 token.")
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: conf.ClientID})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return web.JsonError(err)
	}
	claimsT := new(json.RawMessage)
	if err := idToken.Claims(&claimsT); err != nil {
		return web.JsonError(err)
	}
	buff := new(bytes.Buffer)
	if err := json.Indent(buff, *claimsT, "", "  "); err != nil {
		return web.JsonError(err)
	}
	oidcClaims := &struct {
		IDToken       string `json:"id_token,omitempty"`
		Email         string `json:"email,omitempty"`
		Name          string `json:"name,omitempty"`
		Nickname      string `json:"nickname,omitempty"`
		RefreshToken  string `json:"refresh_token,omitempty"`
		UserID        string `json:"user_id,omitempty"`
		Username      string `json:"username,omitempty"`
		EmployeeNO    string `json:"employee_no,omitempty"`
		EmailVerified bool   `json:"email_verified,omitempty"`
		LoginType     string `json:"login_type"`
	}{}
	if err := json.Unmarshal(buff.Bytes(), oidcClaims); err != nil {
		return web.JsonError(err)
	}
	if oidcClaims.Email == "" {
		return web.JsonErrorMsg("Email is required from OIDC provider")
	}
	// 查找或创建用户
	user := services.UserService.FindOne(sqls.NewCnd().Eq("email", oidcClaims.Email))
	if user == nil {
		nickname := oidcClaims.Nickname
		if nickname == "" {
			nickname = oidcClaims.Name
		}
		if nickname == "" {
			nickname = "User"
		}
		user = &models.User{
			Email:         sqls.SqlNullString(oidcClaims.Email),
			EmailVerified: oidcClaims.EmailVerified,
			Nickname:      nickname,
			Username:      sqls.SqlNullString(strs.UUID()), // 生成随机用户名
			Status:        constants.StatusOk,
			CreateTime:    time.Now().Unix(),
			UpdateTime:    time.Now().Unix(),
		}
		if err := services.UserService.Create(user); err != nil {
			return web.JsonError(err)
		}
	}
	// 跳转回原地址
	redirect := "/"
	if r, ok := claims["redirect"].(string); ok && r != "" {
		redirect = r
	}
	return render.BuildLoginSuccess(c.Ctx, user, redirect)
}
