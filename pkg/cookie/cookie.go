package cookie

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// formatDomain formats the domain.
func formatDomain(domain string) string {
	if domain != "localhost" && !strings.HasPrefix(domain, ".") {
		return "." + domain
	}
	return domain
}

// SetCookie sets cookies.
func SetCookie(ctx *gin.Context, accessToken, refreshToken, domain string) {
	formattedDomain := formatDomain(domain)
	ctx.SetCookie("access_token", accessToken, 60*60*24, "/", formattedDomain, true, true)
	ctx.SetCookie("refresh_token", refreshToken, 60*60*24*30, "/", formattedDomain, true, true)
}

// SetRegisterCookie sets registration cookies.
func SetRegisterCookie(ctx *gin.Context, registerToken, domain string) {
	formattedDomain := formatDomain(domain)
	ctx.SetCookie("register_token", registerToken, 60*60, "/", formattedDomain, true, true)
}

// ClearCookie clears cookies.
func ClearCookie(ctx *gin.Context) {
	ctx.SetCookie("access_token", "", -1, "/", "", true, true)
	ctx.SetCookie("refresh_token", "", -1, "/", "", true, true)
}

// ClearRegisterCookie clears registration cookies.
func ClearRegisterCookie(ctx *gin.Context) {
	ctx.SetCookie("register_token", "", -1, "/", "", true, true)
}

// ClearAllCookie clears all cookies.
func ClearAllCookie(ctx *gin.Context) {
	ClearCookie(ctx)
	ClearRegisterCookie(ctx)
}

// GetCookie gets cookies.
func GetCookie(ctx *gin.Context, key string) (string, error) {
	return ctx.Cookie(key)
}

// GetRegisterCookie gets registration cookies.
func GetRegisterCookie(ctx *gin.Context, key string) (string, error) {
	return ctx.Cookie(key)
}
