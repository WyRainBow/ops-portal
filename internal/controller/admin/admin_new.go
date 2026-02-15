package admin

import (
	"github.com/WyRainBow/ops-portal/api/admin"
)

type ControllerV1 struct{}

func NewV1() admin.IAdminV1 {
	return &ControllerV1{}
}

