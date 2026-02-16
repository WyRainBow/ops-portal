// =================================================================================
// Code generated and maintained by GoFrame CLI tool style. Keep stable for routing.
// =================================================================================

package auth

import (
	"context"

	"github.com/WyRainBow/ops-portal/api/auth/v1"
)

type IAuthV1 interface {
	Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error)
	Me(ctx context.Context, req *v1.MeReq) (res *v1.MeRes, err error)
	// Register is optional; keep for compatibility if you later want it.
	Register(ctx context.Context, req *v1.RegisterReq) (res *v1.RegisterRes, err error)
}

