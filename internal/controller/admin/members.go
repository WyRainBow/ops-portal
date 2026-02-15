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

type memberJoinRow struct {
	Member  store.Member `gorm:"embedded"`
	Username *string     `gorm:"column:username"`
	UserRole *string     `gorm:"column:user_role"`
}

func (c *ControllerV1) Members(ctx context.Context, req *v1.MembersReq) (res *v1.MembersRes, err error) {
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

	base := db.WithContext(ctx).Table("members m").
		Select("m.*, u.username as username, u.role as user_role").
		Joins("LEFT JOIN users u ON u.id = m.user_id")

	if kw := strings.TrimSpace(req.Keyword); kw != "" {
		like := "%" + kw + "%"
		base = base.Where("m.name ILIKE ? OR u.username ILIKE ?", like, like)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, gerror.Newf("db count failed: %v", err)
	}

	var rows []memberJoinRow
	if err := base.Order("m.updated_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows).Error; err != nil {
		return nil, gerror.Newf("db query failed: %v", err)
	}

	items := make([]v1.MemberItem, 0, len(rows))
	for _, r := range rows {
		username := ""
		userRole := ""
		if r.Username != nil {
			username = *r.Username
		}
		if r.UserRole != nil {
			userRole = *r.UserRole
		}
		pos := ""
		team := ""
		if r.Member.Position != nil {
			pos = *r.Member.Position
		}
		if r.Member.Team != nil {
			team = *r.Member.Team
		}
		created := ""
		updated := ""
		if r.Member.CreatedAt != nil {
			created = r.Member.CreatedAt.UTC().Format(time.RFC3339Nano)
		}
		if r.Member.UpdatedAt != nil {
			updated = r.Member.UpdatedAt.UTC().Format(time.RFC3339Nano)
		}
		items = append(items, v1.MemberItem{
			ID:       r.Member.ID,
			Name:     r.Member.Name,
			Username: username,
			Position: pos,
			Team:     team,
			Status:   r.Member.Status,
			UserID:   r.Member.UserID,
			UserRole: userRole,
			CreatedAt: created,
			UpdatedAt: updated,
		})
	}

	return &v1.MembersRes{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (c *ControllerV1) CreateMember(ctx context.Context, req *v1.CreateMemberReq) (res *v1.CreateMemberRes, err error) {
	operator, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if req.UserRole != "" && req.UserRole != "admin" && req.UserRole != "member" && req.UserRole != "user" {
		return nil, gerror.New("不支持的角色")
	}
	status := req.Status
	if strings.TrimSpace(status) == "" {
		status = "active"
	}

	db, err := store.DB(ctx)
	if err != nil {
		return nil, gerror.Newf("db init failed: %v", err)
	}

	var u store.User
	if err := db.WithContext(ctx).First(&u, "id = ?", req.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gerror.New("关联用户不存在")
		}
		return nil, gerror.Newf("db query failed: %v", err)
	}
	var exists store.Member
	if err := db.WithContext(ctx).First(&exists, "user_id = ?", req.UserID).Error; err == nil {
		return nil, gerror.New("该用户已是成员")
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return nil, gerror.Newf("db query failed: %v", err)
	}

	m := store.Member{
		Name:   u.Username,
		Email: nil,
		Status: status,
		UserID: &req.UserID,
	}
	if strings.TrimSpace(req.Position) != "" {
		m.Position = ptr(req.Position)
	}
	if strings.TrimSpace(req.Team) != "" {
		m.Team = ptr(req.Team)
	}

	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if req.UserRole != "" && req.UserRole != u.Role {
			old := u.Role
			newR := req.UserRole
			a := store.PermissionAuditLog{
				OperatorUserID: &operator.ID,
				TargetUserID:   &u.ID,
				FromRole:       &old,
				ToRole:         &newR,
				Action:         "update_role_from_member",
			}
			if err := tx.Create(&a).Error; err != nil {
				return err
			}
			if err := tx.Model(&store.User{}).Where("id = ?", u.ID).Updates(map[string]any{"role": newR, "updated_at": time.Now().UTC()}).Error; err != nil {
				return err
			}
			u.Role = newR
		}
		return tx.Create(&m).Error
	}); err != nil {
		return nil, gerror.Newf("db write failed: %v", err)
	}

	return &v1.MemberItem{
		ID:       m.ID,
		Name:     m.Name,
		Username: u.Username,
		Position: deref(m.Position),
		Team:     deref(m.Team),
		Status:   m.Status,
		UserID:   m.UserID,
		UserRole: u.Role,
	}, nil
}

func (c *ControllerV1) UpdateMember(ctx context.Context, req *v1.UpdateMemberReq) (res *v1.UpdateMemberRes, err error) {
	operator, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if req.UserRole != "" && req.UserRole != "admin" && req.UserRole != "member" && req.UserRole != "user" {
		return nil, gerror.New("不支持的角色")
	}
	status := req.Status
	if strings.TrimSpace(status) == "" {
		status = "active"
	}

	db, err := store.DB(ctx)
	if err != nil {
		return nil, gerror.Newf("db init failed: %v", err)
	}

	var m store.Member
	if err := db.WithContext(ctx).First(&m, "id = ?", req.MemberID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gerror.New("成员不存在")
		}
		return nil, gerror.Newf("db query failed: %v", err)
	}

	_ = operator

	var u store.User
	if err := db.WithContext(ctx).First(&u, "id = ?", req.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gerror.New("关联用户不存在")
		}
		return nil, gerror.Newf("db query failed: %v", err)
	}
	// prevent duplicate mapping
	var dup store.Member
	if err := db.WithContext(ctx).Where("user_id = ? AND id <> ?", req.UserID, req.MemberID).First(&dup).Error; err == nil {
		return nil, gerror.New("该用户已是其他成员")
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return nil, gerror.Newf("db query failed: %v", err)
	}

	updates := map[string]any{
		"name":       u.Username,
		"email":      nil,
		"position":   nil,
		"team":       nil,
		"status":     status,
		"user_id":    req.UserID,
		"updated_at": time.Now().UTC(),
	}
	if strings.TrimSpace(req.Position) != "" {
		updates["position"] = req.Position
	}
	if strings.TrimSpace(req.Team) != "" {
		updates["team"] = req.Team
	}

	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if req.UserRole != "" && req.UserRole != u.Role {
			old := u.Role
			newR := req.UserRole
			a := store.PermissionAuditLog{
				OperatorUserID: &operator.ID,
				TargetUserID:   &u.ID,
				FromRole:       &old,
				ToRole:         &newR,
				Action:         "update_role_from_member",
			}
			if err := tx.Create(&a).Error; err != nil {
				return err
			}
			if err := tx.Model(&store.User{}).Where("id = ?", u.ID).Updates(map[string]any{"role": newR, "updated_at": time.Now().UTC()}).Error; err != nil {
				return err
			}
			u.Role = newR
		}
		return tx.Model(&store.Member{}).Where("id = ?", req.MemberID).Updates(updates).Error
	}); err != nil {
		return nil, gerror.Newf("db write failed: %v", err)
	}

	return &v1.MemberItem{
		ID:       m.ID,
		Name:     u.Username,
		Username: u.Username,
		Position: req.Position,
		Team:     req.Team,
		Status:   status,
		UserID:   &req.UserID,
		UserRole: u.Role,
	}, nil
}

func (c *ControllerV1) DeleteMember(ctx context.Context, req *v1.DeleteMemberReq) (res *v1.DeleteMemberRes, err error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	db, err := store.DB(ctx)
	if err != nil {
		return nil, gerror.Newf("db init failed: %v", err)
	}

	if err := db.WithContext(ctx).Delete(&store.Member{}, "id = ?", req.MemberID).Error; err != nil {
		return nil, gerror.Newf("db delete failed: %v", err)
	}
	return &v1.DeleteMemberRes{Success: true}, nil
}
