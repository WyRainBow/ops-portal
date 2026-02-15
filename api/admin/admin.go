// =================================================================================
// Admin API interface for GoFrame router binding.
// =================================================================================

package admin

import (
	"context"

	"github.com/WyRainBow/ops-portal/api/admin/v1"
)

type IAdminV1 interface {
	Overview(ctx context.Context, req *v1.OverviewReq) (res *v1.OverviewRes, err error)

	ApiRoutes(ctx context.Context, req *v1.ApiRoutesReq) (res *v1.ApiRoutesRes, err error)

	Users(ctx context.Context, req *v1.UsersReq) (res *v1.UsersRes, err error)
	UpdateUserRole(ctx context.Context, req *v1.UpdateUserRoleReq) (res *v1.UserItem, err error)
	UpdateUserQuota(ctx context.Context, req *v1.UpdateUserQuotaReq) (res *v1.UserItem, err error)

	Members(ctx context.Context, req *v1.MembersReq) (res *v1.MembersRes, err error)
	CreateMember(ctx context.Context, req *v1.CreateMemberReq) (res *v1.MemberItem, err error)
	UpdateMember(ctx context.Context, req *v1.UpdateMemberReq) (res *v1.MemberItem, err error)
	DeleteMember(ctx context.Context, req *v1.DeleteMemberReq) (res *v1.DeleteMemberRes, err error)

	PermissionRoles(ctx context.Context, req *v1.PermissionRolesReq) (res *v1.PermissionRolesRes, err error)
	PermissionAudits(ctx context.Context, req *v1.PermissionAuditsReq) (res *v1.PermissionAuditsRes, err error)

	RequestLogs(ctx context.Context, req *v1.RequestLogsReq) (res *v1.RequestLogsRes, err error)
	ErrorLogs(ctx context.Context, req *v1.ErrorLogsReq) (res *v1.ErrorLogsRes, err error)

	Traces(ctx context.Context, req *v1.TracesReq) (res *v1.TracesRes, err error)
	TraceDetail(ctx context.Context, req *v1.TraceDetailReq) (res *v1.TraceDetailRes, err error)

	RuntimeStatus(ctx context.Context, req *v1.RuntimeStatusReq) (res *v1.RuntimeStatusRes, err error)
	RuntimeLogs(ctx context.Context, req *v1.RuntimeLogsReq) (res *v1.RuntimeLogsRes, err error)
}
