package admin

import (
	"context"
	"time"

	v1 "github.com/WyRainBow/ops-portal/api/admin/v1"
	"github.com/WyRainBow/ops-portal/internal/store"

	"github.com/gogf/gf/v2/errors/gerror"
)

var roleMatrix = map[string]any{
	"admin": map[string]any{
		"users":   []string{"read", "write", "grant_admin"},
		"members": []string{"read", "write"},
		"logs":    []string{"read"},
		"traces":  []string{"read"},
	},
	"member": map[string]any{
		"users":   []string{"read", "write_non_admin"},
		"members": []string{"read", "write"},
		"logs":    []string{"read"},
		"traces":  []string{"read"},
	},
	"user": map[string]any{
		"users":   []string{},
		"members": []string{},
		"logs":    []string{},
		"traces":  []string{},
	},
}

func (c *ControllerV1) PermissionRoles(ctx context.Context, req *v1.PermissionRolesReq) (res *v1.PermissionRolesRes, err error) {
	if _, err := requireAdminOrMember(ctx); err != nil {
		return nil, err
	}
	return &v1.PermissionRolesRes{Roles: roleMatrix}, nil
}

type auditJoinRow struct {
	store.PermissionAuditLog
	OperatorUsername *string `gorm:"column:operator_username"`
	TargetUsername   *string `gorm:"column:target_username"`
}

func (c *ControllerV1) PermissionAudits(ctx context.Context, req *v1.PermissionAuditsReq) (res *v1.PermissionAuditsRes, err error) {
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

	base := db.WithContext(ctx).Table("permission_audit_logs p").
		Select("p.*, ou.username as operator_username, tu.username as target_username").
		Joins("LEFT JOIN users ou ON ou.id = p.operator_user_id").
		Joins("LEFT JOIN users tu ON tu.id = p.target_user_id")

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, gerror.Newf("db count failed: %v", err)
	}

	var rows []auditJoinRow
	if err := base.Order("p.created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows).Error; err != nil {
		return nil, gerror.Newf("db query failed: %v", err)
	}

	items := make([]v1.PermissionAuditItem, 0, len(rows))
	for _, r := range rows {
		created := ""
		if r.CreatedAt != nil {
			created = r.CreatedAt.UTC().Format(time.RFC3339Nano)
		}
		opName := ""
		tName := ""
		if r.OperatorUsername != nil {
			opName = *r.OperatorUsername
		}
		if r.TargetUsername != nil {
			tName = *r.TargetUsername
		}
		from := ""
		to := ""
		if r.FromRole != nil {
			from = *r.FromRole
		}
		if r.ToRole != nil {
			to = *r.ToRole
		}
		items = append(items, v1.PermissionAuditItem{
			ID:               r.ID,
			OperatorUserID:   r.OperatorUserID,
			TargetUserID:     r.TargetUserID,
			OperatorUsername: opName,
			TargetUsername:   tName,
			FromRole:         from,
			ToRole:           to,
			Action:           r.Action,
			CreatedAt:        created,
		})
	}

	return &v1.PermissionAuditsRes{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

