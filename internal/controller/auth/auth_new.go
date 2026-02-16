package auth

import (
	"github.com/WyRainBow/ops-portal/api/auth"
)

// ControllerV1 implements /api/auth/* endpoints.
type ControllerV1 struct{}

func NewV1() auth.IAuthV1 {
	return &ControllerV1{}
}

