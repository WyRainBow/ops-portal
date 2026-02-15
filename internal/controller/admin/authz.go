package admin

import (
	"context"
	"strings"

	"github.com/WyRainBow/ops-portal/internal/security"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

type AuthUser struct {
	ID       int64
	Username string
	Role     string
}

func currentUserFromJWT(ctx context.Context) (*AuthUser, error) {
	r := g.RequestFromCtx(ctx)
	if r == nil {
		return nil, gerror.New("未提供有效的认证信息")
	}
	authz := r.Header.Get("Authorization")
	if !strings.HasPrefix(strings.ToLower(authz), "bearer ") {
		return nil, gerror.New("未提供有效的认证信息")
	}
	token := strings.TrimSpace(authz[7:])
	claims, err := security.ParseToken(token)
	if err != nil {
		return nil, gerror.New("Token 无效或已过期")
	}
	uid, _ := parseInt64(claims.Subject)
	return &AuthUser{
		ID:       uid,
		Username: claims.Username,
		Role:     claims.Role,
	}, nil
}

func requireAdminOrMember(ctx context.Context) (*AuthUser, error) {
	u, err := currentUserFromJWT(ctx)
	if err != nil {
		return nil, err
	}
	if u.Role != "admin" && u.Role != "member" {
		return nil, gerror.New("仅管理员或成员可访问")
	}
	return u, nil
}

func requireAdmin(ctx context.Context) (*AuthUser, error) {
	u, err := currentUserFromJWT(ctx)
	if err != nil {
		return nil, err
	}
	if u.Role != "admin" {
		return nil, gerror.New("仅管理员可访问")
	}
	return u, nil
}

func parseInt64(s string) (int64, error) {
	var n int64
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, gerror.New("invalid int")
		}
		n = n*10 + int64(ch-'0')
	}
	return n, nil
}

