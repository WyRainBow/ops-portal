package auth

import (
	"context"
	"os"
	"strings"

	v1 "github.com/WyRainBow/ops-portal/api/auth/v1"
	"github.com/WyRainBow/ops-portal/internal/security"
	"github.com/WyRainBow/ops-portal/internal/store"

	"github.com/gogf/gf/v2/errors/gerror"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (c *ControllerV1) Register(ctx context.Context, req *v1.RegisterReq) (res *v1.RegisterRes, err error) {
	if os.Getenv("OPS_PORTAL_DISABLE_REGISTER") == "true" {
		return nil, gerror.New("注册已禁用")
	}
	db, err := store.DB(ctx)
	if err != nil {
		return nil, gerror.Newf("db init failed: %v", err)
	}

	username := strings.TrimSpace(req.Username)
	if username == "" || req.Password == "" {
		return nil, gerror.New("账号或密码不能为空")
	}

	// Check duplicates.
	var exists store.User
	if err := db.WithContext(ctx).Model(&store.User{}).Where("username = ? OR email = ?", username, username).First(&exists).Error; err == nil {
		return nil, gerror.New("该账号已注册")
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return nil, gerror.Newf("db query failed: %v", err)
	}

	h, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, gerror.Newf("password hash failed: %v", err)
	}

	u := store.User{
		Username:     username,
		Email:        nil,
		PasswordHash: string(h),
		Role:         "user",
	}
	// If username looks like email, store it.
	if strings.Contains(username, "@") {
		u.Email = &username
	}
	if err := db.WithContext(ctx).Create(&u).Error; err != nil {
		return nil, gerror.Newf("db insert failed: %v", err)
	}

	token, err := security.CreateToken(u.ID, u.Username, u.Role)
	if err != nil {
		return nil, gerror.Newf("token create failed: %v", err)
	}

	return &v1.RegisterRes{
		AccessToken: token,
		TokenType:   "bearer",
		User: v1.UserSummary{
			ID:       u.ID,
			Username: u.Username,
			Email:    deref(u.Email),
			Role:     u.Role,
		},
	}, nil
}
