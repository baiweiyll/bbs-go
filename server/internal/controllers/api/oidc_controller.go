package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/config"
	"bbs-go/internal/services"

	"github.com/coreos/go-oidc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
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
			"groups",
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
		context.CookieHTTPOnly(true),
		context.CookieExpires(10*time.Minute),
		context.CookiePath("/"),
		context.CookieDomain(""),
	)
	url := oauth2Config.AuthCodeURL(state)
	if c.Ctx.FormValue("type") != "" {
		url += "&type=" + c.Ctx.FormValue("type")
	}
	slog.Debug("OIDC login", slog.Any("state", state), slog.Any("redirect", redirect), slog.Any("URL", url))
	c.Ctx.Redirect(url, iris.StatusFound)
	return nil
}

// GetCallback OIDC 登录回调
func (c *OIDCController) GetCallback() *web.JsonResult {
	conf := config.Instance.OIDC
	incomingState := c.Ctx.URLParam("state")
	if incomingState == "" {
		slog.Error("OIDC state not found")
		c.redirectWithError(conf.Console, c.Ctx, fmt.Errorf("OIDC state not found"))
		return nil
	}
	stateToken := c.Ctx.GetCookie("oidc_state_token")
	if stateToken == "" {
		slog.Error("OIDC state token not found")
		c.redirectWithError(conf.Console, c.Ctx, fmt.Errorf("OIDC state token not found"))
		return nil
	}
	// 验证 State Token
	token, err := jwt.Parse(stateToken, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(conf.SecretKey), nil
	})
	if err != nil || !token.Valid {
		slog.Error("OIDC state token invalid", slog.Any("error", err))
		c.redirectWithError(conf.Console, c.Ctx, fmt.Errorf("invalid OIDC state token"))
		return nil
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		slog.Error("invalid OIDC state token clasims")
		c.redirectWithError(conf.Console, c.Ctx, fmt.Errorf("invalid OIDC state token claims"))
		return nil
	}
	expectedState := claims["state"].(string)
	if expectedState != incomingState {
		slog.Error("OIDC state mismatch")
		c.redirectWithError(conf.Console, c.Ctx, fmt.Errorf("OIDC state mismatch"))
		return nil
	}
	ctx := c.Ctx.Request().Context()
	provider, err := oidc.NewProvider(ctx, conf.Issuer)
	if err != nil {
		slog.Error("OIDC provider error", slog.Any("error", err))
		c.redirectWithError(conf.Console, c.Ctx, err)
		return nil
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
		slog.Error("OIDC exchange error", slog.Any("error", err))
		c.redirectWithError(conf.Console, c.Ctx, err)
		return nil
	}
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		slog.Error("OIDC no id_token field in oauth2 token")
		c.redirectWithError(conf.Console, c.Ctx, fmt.Errorf("no id_token field in oauth2 token"))
		return nil
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: conf.ClientID})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		slog.Error("OIDC id_token error", slog.Any("error", err))
		c.redirectWithError(conf.Console, c.Ctx, err)
		return nil
	}
	claimsT := new(json.RawMessage)
	if err := idToken.Claims(&claimsT); err != nil {
		slog.Error("OIDC claims token error", slog.Any("error", err))
		c.redirectWithError(conf.Console, c.Ctx, err)
		return nil
	}
	buff := new(bytes.Buffer)
	if err := json.Indent(buff, *claimsT, "", "  "); err != nil {
		slog.Error("OIDC indent error", slog.Any("error", err))
		c.redirectWithError(conf.Console, c.Ctx, err)
		return nil
	}
	oidcClaims := &struct {
		IDToken       string   `json:"id_token,omitempty"`
		Email         string   `json:"email,omitempty"`
		Name          string   `json:"name,omitempty"`
		Nickname      string   `json:"nickname,omitempty"`
		RefreshToken  string   `json:"refresh_token,omitempty"`
		UserID        string   `json:"user_id,omitempty"`
		Username      string   `json:"username,omitempty"`
		EmployeeNO    string   `json:"employee_no,omitempty"`
		EmailVerified bool     `json:"email_verified,omitempty"`
		LoginType     string   `json:"login_type"`
		Groups        []string `json:"groups,omitempty"`
	}{}
	if err := json.Unmarshal(buff.Bytes(), oidcClaims); err != nil {
		slog.Error("OIDC parse error", slog.Any("error", err))
		c.redirectWithError(conf.Console, c.Ctx, err)
		return nil
	}
	if oidcClaims.Email == "" {
		slog.Error("OIDC email is required")
		c.redirectWithError(conf.Console, c.Ctx, fmt.Errorf("email is required from OIDC provider"))
		return nil
	}
	if oidcClaims.Groups == nil {
		slog.Error("No grouping information")
		c.redirectWithError(conf.Console, c.Ctx, fmt.Errorf("You do not have access permission"))
		return nil
	} else {
		hasBbsGoGroup := false
		for _, group := range oidcClaims.Groups {
			if strings.Compare("bbs-go", group) == 0 {
				hasBbsGoGroup = true
				break
			}
		}
		if !hasBbsGoGroup {
			slog.Error("No valid grouping information")
			c.redirectWithError(conf.Console, c.Ctx, fmt.Errorf("You do not have permission to access"))
			return nil
		}
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
			slog.Error("OIDC create user error", slog.Any("error", err))
			c.redirectWithError(conf.Console, c.Ctx, err)
			return nil
		}
	}
	res := struct {
		Redirect string      `json:"redirect"`
		Token    string      `json:"token"`
		User     models.User `json:"user"`
	}{}
	// 跳转回原地址
	redirect := "/"
	if r, ok := claims["redirect"].(string); ok && r != "" {
		redirect = r
	}
	res.Redirect = redirect
	res.User = *user
	tokenGenerate, err := services.UserTokenService.Generate(user.Id)
	if err != nil {
		slog.Error("OIDC generate token error", slog.Any("error", err))
		c.redirectWithError(conf.Console, c.Ctx, err)
		return nil
	}
	res.Token = tokenGenerate
	c.Ctx.SetCookieKV(
		constants.CookieTokenKey,
		tokenGenerate,
		context.CookieHTTPOnly(true),
		context.CookieExpires(365*24*time.Hour),
		context.CookieDomain(conf.Domain),
	)
	data, err := json.Marshal(res)
	if err != nil {
		slog.Error("Failed to marshal user to json", slog.Any("error", err))
		c.redirectWithError(conf.Console, c.Ctx, err)
		return nil
	}
	base64Data := base64.RawURLEncoding.EncodeToString(data)
	url := fmt.Sprintf("%s?data=%s", conf.Console, base64Data)
	slog.Debug("Callback redirect url", slog.Any("Data", res), slog.Any("URL", url))
	c.Ctx.Redirect(url, iris.StatusFound)
	return nil
}

// 退出登录
func (c *OIDCController) GetSignout() *web.JsonResult {
	config := config.Instance.OIDC
	err := services.UserTokenService.Signout(c.Ctx)
	if err != nil {
		return web.JsonError(err)
	}
	c.Ctx.RemoveCookie(
		"oidc_state_token",
		context.CookieHTTPOnly(true),
		context.CookiePath("/"),
		context.CookieDomain(""),
	)
	url := fmt.Sprintf("%s/logout?client_id=%s&redirect_uri=%s",
		config.Issuer,
		config.ClientID,
		config.Redirect)
	slog.Debug("Logout redirect", slog.Any("URL", url))
	c.Ctx.Redirect(url, iris.StatusFound)
	return nil
}

func (c *OIDCController) redirectWithError(console string, ctx iris.Context, err error) {
	errResponse := struct {
		ErrorCode int         `json:"errorCode"`
		Message   string      `json:"message"`
		Data      interface{} `json:"data,omitempty"`
		Success   bool        `json:"success"`
	}{
		ErrorCode: 0,
		Message:   err.Error(),
		Success:   false,
	}
	data, _ := json.Marshal(errResponse)
	encoded := base64.RawURLEncoding.EncodeToString(data)
	target := fmt.Sprintf("%s?data=%s", console, encoded)
	ctx.Redirect(target, iris.StatusFound)
}
