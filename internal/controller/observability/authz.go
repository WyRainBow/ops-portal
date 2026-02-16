package observability

import (
	"context"
	"strings"

	"github.com/WyRainBow/ops-portal/internal/security"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

func requireAdminOrMember(ctx context.Context) error {
	r := g.RequestFromCtx(ctx)
	if r == nil {
		return gerror.New("未提供有效的认证信息")
	}
	authz := r.Header.Get("Authorization")
	if !strings.HasPrefix(strings.ToLower(authz), "bearer ") {
		return gerror.New("未提供有效的认证信息")
	}
	token := strings.TrimSpace(authz[7:])
	claims, err := security.ParseToken(token)
	if err != nil {
		return gerror.New("Token 无效或已过期")
	}
	if claims.Role != "admin" && claims.Role != "member" {
		return gerror.New("仅管理员或成员可访问")
	}
	return nil
}

