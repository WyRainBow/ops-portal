package v1

import "github.com/gogf/gf/v2/frame/g"

type RegisterReq struct {
	g.Meta   `path:"/auth/register" method:"post" summary:"注册"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRes struct {
	AccessToken string      `json:"access_token"`
	TokenType   string      `json:"token_type"`
	User        UserSummary `json:"user"`
}

type LoginReq struct {
	g.Meta   `path:"/auth/login" method:"post" summary:"登录"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRes struct {
	AccessToken string      `json:"access_token"`
	TokenType   string      `json:"token_type"`
	User        UserSummary `json:"user"`
}

type MeReq struct {
	g.Meta `path:"/auth/me" method:"get" summary:"当前用户"`
}

type MeRes struct {
	UserSummary
}

type UserSummary struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email,omitempty"`
	Role     string `json:"role,omitempty"`
}

