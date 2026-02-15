package admin

import (
	"context"
	"strings"
	"time"

	v1 "github.com/WyRainBow/ops-portal/api/admin/v1"
	"github.com/WyRainBow/ops-portal/internal/store"

	"github.com/gogf/gf/v2/errors/gerror"
	"gorm.io/gorm"
)

func (c *ControllerV1) Users(ctx context.Context, req *v1.UsersReq) (res *v1.UsersRes, err error) {
	if _, err := requireAdminOrMember(ctx); err != nil {
		return nil, err
	}
	db, err := store.DB(ctx)
	if err != nil {
		return nil, gerror.Newf("db init failed: %v", err)
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}

	q := db.WithContext(ctx).Model(&store.User{})
	if kw := strings.TrimSpace(req.Keyword); kw != "" {
		like := "%" + kw + "%"
		q = q.Where("username ILIKE ? OR email ILIKE ?", like, like)
	}
	if role := strings.TrimSpace(req.Role); role != "" {
		q = q.Where("role = ?", role)
	}
	if ip := strings.TrimSpace(req.IP); ip != "" {
		q = q.Where("last_login_ip = ?", ip)
	}

	var total int64
	if req.WithTotal {
		if err := q.Count(&total).Error; err != nil {
			return nil, gerror.Newf("db count failed: %v", err)
		}
	}

	var rows []store.User
	if err := q.Order("updated_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, gerror.Newf("db query failed: %v", err)
	}

	items := make([]v1.UserItem, 0, len(rows))
	for _, r := range rows {
		ip := ""
		if r.LastLoginIP != nil {
			ip = *r.LastLoginIP
		}
		email := ""
		if r.Email != nil {
			email = *r.Email
		}
		created := ""
		updated := ""
		if r.CreatedAt != nil {
			created = r.CreatedAt.UTC().Format(time.RFC3339Nano)
		}
		if r.UpdatedAt != nil {
			updated = r.UpdatedAt.UTC().Format(time.RFC3339Nano)
		}
		items = append(items, v1.UserItem{
			ID:          r.ID,
			Username:    r.Username,
			Email:       email,
			Role:        r.Role,
			LastLoginIP: ip,
			APIQuota:    r.APIQuota,
			CreatedAt:   created,
			UpdatedAt:   updated,
		})
	}

	return &v1.UsersRes{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (c *ControllerV1) UpdateUserRole(ctx context.Context, req *v1.UpdateUserRoleReq) (res *v1.UserItem, err error) {
	operator, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	newRole := strings.ToLower(strings.TrimSpace(req.Role))
	if newRole != "admin" && newRole != "member" && newRole != "user" {
		return nil, gerror.New("不支持的角色")
	}

	db, err := store.DB(ctx)
	if err != nil {
		return nil, gerror.Newf("db init failed: %v", err)
	}

	var u store.User
	if err := db.WithContext(ctx).First(&u, "id = ?", req.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gerror.New("用户不存在")
		}
		return nil, gerror.Newf("db query failed: %v", err)
	}

	oldRole := u.Role
	if oldRole != newRole {
		now := time.Now().UTC()
		if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&store.User{}).Where("id = ?", u.ID).Updates(map[string]any{
				"role":       newRole,
				"updated_at": now,
			}).Error; err != nil {
				return err
			}
			a := store.PermissionAuditLog{
				OperatorUserID: &operator.ID,
				TargetUserID:   &u.ID,
				FromRole:       &oldRole,
				ToRole:         &newRole,
				Action:         "update_role",
			}
			return tx.Create(&a).Error
		}); err != nil {
			return nil, gerror.Newf("db update failed: %v", err)
		}
		u.Role = newRole
	}

	return &v1.UserItem{
		ID:       u.ID,
		Username: u.Username,
		Email:    deref(u.Email),
		Role:     u.Role,
		APIQuota: u.APIQuota,
	}, nil
}

func (c *ControllerV1) UpdateUserQuota(ctx context.Context, req *v1.UpdateUserQuotaReq) (res *v1.UserItem, err error) {
	operator, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}

	db, err := store.DB(ctx)
	if err != nil {
		return nil, gerror.Newf("db init failed: %v", err)
	}

	var u store.User
	if err := db.WithContext(ctx).First(&u, "id = ?", req.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gerror.New("用户不存在")
		}
		return nil, gerror.Newf("db query failed: %v", err)
	}

	now := time.Now().UTC()
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&store.User{}).Where("id = ?", u.ID).Updates(map[string]any{
			"api_quota":  req.APIQuota,
			"updated_at": now,
		}).Error; err != nil {
			return err
		}
		role := u.Role
		a := store.PermissionAuditLog{
			OperatorUserID: &operator.ID,
			TargetUserID:   &u.ID,
			FromRole:       &role,
			ToRole:         &role,
			Action:         "update_quota",
		}
		return tx.Create(&a).Error
	}); err != nil {
		return nil, gerror.Newf("db update failed: %v", err)
	}

	u.APIQuota = req.APIQuota
	return &v1.UserItem{
		ID:       u.ID,
		Username: u.Username,
		Email:    deref(u.Email),
		Role:     u.Role,
		APIQuota: u.APIQuota,
	}, nil
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
