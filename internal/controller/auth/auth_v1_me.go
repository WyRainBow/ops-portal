package auth

import (
	"context"
	"strings"

	v1 "github.com/WyRainBow/ops-portal/api/auth/v1"
	"github.com/WyRainBow/ops-portal/internal/security"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

func bearerTokenFromReq(ctx context.Context) string {
	r := g.RequestFromCtx(ctx)
	if r == nil {
		return ""
	}
	authz := r.Header.Get("Authorization")
	if authz == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(authz), "bearer ") {
		return strings.TrimSpace(authz[7:])
	}
	return ""
}

func (c *ControllerV1) Me(ctx context.Context, req *v1.MeReq) (res *v1.MeRes, err error) {
	token := bearerTokenFromReq(ctx)
	if token == "" {
		return nil, gerror.New("未提供有效的认证信息")
	}
	claims, err := security.ParseToken(token)
	if err != nil {
		return nil, gerror.New("Token 无效或已过期")
	}

	// We intentionally return claim-based info (same as Resume-Agent /api/admin/* auth optimization).
	// If you want stronger consistency, add DB lookup here later.
	userID := int64(0)
	if claims.Subject != "" {
		// ignore parse error -> keep 0
		if v, perr := parseInt64(claims.Subject); perr == nil {
			userID = v
		}
	}

	return &v1.MeRes{
		UserSummary: v1.UserSummary{
			ID:       userID,
			Username: claims.Username,
			Role:     claims.Role,
		},
	}, nil
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

