package middleware

import (
	"context"
	"strings"

	"github.com/WyRainBow/ops-portal/internal/security"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// AuthConfig defines the configuration for JWT authentication middleware.
type AuthConfig struct {
	// SkipPaths are paths that skip authentication (e.g., /api/auth/login, /api/auth/register)
	SkipPaths []string

	// SkipPrefixes are path prefixes that skip authentication (e.g., /api/auth)
	SkipPrefixes []string

	// RequireRole can optionally specify required role (empty means any authenticated user)
	RequireRole string
}

// DefaultSkipPaths are paths that should always skip authentication.
var DefaultSkipPaths = []string{
	"/api/auth/login",
	"/api/auth/register",
	"/api/health",
	"/api/swagger",
	"/api/openapi",
}

// JWTAuth creates a JWT authentication middleware.
// It verifies the JWT token from Authorization header and sets user context.
func JWTAuth(config *AuthConfig) func(r *ghttp.Request) {
	if config == nil {
		config = &AuthConfig{}
	}

	// Merge with default skip paths
	skipPathMap := make(map[string]bool)
	for _, p := range DefaultSkipPaths {
		skipPathMap[p] = true
	}
	for _, p := range config.SkipPaths {
		skipPathMap[p] = true
	}

	skipPrefixMap := make(map[string]bool)
	for _, p := range config.SkipPrefixes {
		skipPrefixMap[p] = true
	}

	return func(r *ghttp.Request) {
		path := r.URL.Path

		// Check if path should skip authentication
		if skipPathMap[path] {
			r.Middleware.Next()
			return
		}

		// Check if path prefix should skip authentication
		for _, prefix := range config.SkipPrefixes {
			if strings.HasPrefix(path, prefix) {
				r.Middleware.Next()
				return
			}
		}

		// Extract Authorization header
		authz := r.Header.Get("Authorization")
		if authz == "" {
			r.Response.WriteJson(AuthErrorResponse("未提供认证信息"))
			r.Exit()
			return
		}

		// Check Bearer prefix
		if !strings.HasPrefix(strings.ToLower(authz), "bearer ") {
			r.Response.WriteJson(AuthErrorResponse("认证格式错误"))
			r.Exit()
			return
		}

		token := strings.TrimSpace(authz[7:])
		claims, err := security.ParseToken(token)
		if err != nil {
			r.Response.WriteJson(AuthErrorResponse("Token 无效或已过期"))
			r.Exit()
			return
		}

		// Check role requirement
		if config.RequireRole != "" && claims.Role != config.RequireRole {
			r.Response.WriteJson(AuthErrorResponse("权限不足"))
			r.Exit()
			return
		}

		// Set user context for downstream handlers
		ctx := SetUserContext(r.Context(), &UserContext{
			UserID:   claims.Subject,
			Username: claims.Username,
			Role:     claims.Role,
		})
		r.SetCtx(ctx)

		r.Middleware.Next()
	}
}

// UserContext holds authenticated user information.
type UserContext struct {
	UserID   string
	Username string
	Role     string
}

// userContextKey is the key used to store user context in request context.
type userContextKey struct{}

var userKey = userContextKey{}

// SetUserContext sets the user context in the given context.
func SetUserContext(ctx context.Context, user *UserContext) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// GetUserContext retrieves the user context from the given context.
// Returns nil if no user context is set.
func GetUserContext(ctx context.Context) *UserContext {
	if user, ok := ctx.Value(userKey).(*UserContext); ok {
		return user
	}
	return nil
}

// RequireRole checks if the current user has the required role.
// Returns gerror if user doesn't have the required role.
func RequireRole(ctx context.Context, role string) error {
	user := GetUserContext(ctx)
	if user == nil {
		return gerror.New("未提供有效的认证信息")
	}
	if user.Role != role {
		return gerror.New("权限不足")
	}
	return nil
}

// RequireAnyRole checks if the current user has any of the required roles.
// Returns gerror if user doesn't have any of the required roles.
func RequireAnyRole(ctx context.Context, roles ...string) error {
	user := GetUserContext(ctx)
	if user == nil {
		return gerror.New("未提供有效的认证信息")
	}
	for _, role := range roles {
		if user.Role == role {
			return nil
		}
	}
	return gerror.New("权限不足")
}

// AuthErrorResponse creates a standardized error response for authentication failures.
func AuthErrorResponse(msg string) map[string]any {
	return map[string]any{
		"success": false,
		"message": msg,
		"code":    "UNAUTHORIZED",
	}
}

// AdminAuth creates a middleware that requires admin role.
func AdminAuth() func(r *ghttp.Request) {
	return JWTAuth(&AuthConfig{
		RequireRole: "admin",
	})
}

// AdminOrMemberAuth creates a middleware that requires admin or member role.
func AdminOrMemberAuth() func(r *ghttp.Request) {
	// This will be used with RequireAnyRole in handlers
	return JWTAuth(&AuthConfig{})
}
