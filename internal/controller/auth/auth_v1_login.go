package auth

import (
	"context"
	"strings"
	"time"

	v1 "github.com/WyRainBow/ops-portal/api/auth/v1"
	"github.com/WyRainBow/ops-portal/internal/security"
	"github.com/WyRainBow/ops-portal/internal/store"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"gorm.io/gorm"
)

func (c *ControllerV1) Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error) {
	db, err := store.DB(ctx)
	if err != nil {
		return nil, gerror.Newf("db init failed: %v", err)
	}

	login := strings.TrimSpace(req.Username)
	if login == "" || req.Password == "" {
		return nil, gerror.New("账号或密码不能为空")
	}

	var u store.User
	q := db.WithContext(ctx).Model(&store.User{})
	// Prefer single-index lookup: if contains '@' try email first, else username first.
	if strings.Contains(login, "@") {
		err = q.Where("email = ?", login).First(&u).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, gerror.Newf("db query failed: %v", err)
		}
		if err == gorm.ErrRecordNotFound {
			err = q.Where("username = ?", login).First(&u).Error
		}
	} else {
		err = q.Where("username = ?", login).First(&u).Error
		if err == gorm.ErrRecordNotFound {
			err = q.Where("email = ?", login).First(&u).Error
		}
	}
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gerror.New("账号或密码错误")
		}
		return nil, gerror.Newf("db query failed: %v", err)
	}

	ok, verr := security.VerifyPassword(req.Password, u.PasswordHash)
	if verr != nil {
		return nil, gerror.Newf("password verify failed: %v", verr)
	}
	if !ok {
		return nil, gerror.New("账号或密码错误")
	}

	token, err := security.CreateToken(u.ID, u.Username, u.Role)
	if err != nil {
		return nil, gerror.Newf("token create failed: %v", err)
	}

	// Best-effort update last_login_ip, do not fail login on error.
	if r := g.RequestFromCtx(ctx); r != nil {
		ip := r.GetClientIp()
		if ip != "" {
			_ = db.WithContext(ctx).Model(&store.User{}).Where("id = ?", u.ID).Updates(map[string]any{
				"last_login_ip": ip,
				"updated_at":    time.Now().UTC(),
			}).Error
		}
	}

	return &v1.LoginRes{
		AccessToken: token,
		TokenType:   "bearer",
		User: v1.UserSummary{
			ID:       u.ID,
			Username: u.Username,
			Email:    u.Email,
			Role:     u.Role,
		},
	}, nil
}

