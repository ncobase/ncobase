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

// Set sets cookies.
func Set(ctx *gin.Context, accessToken, refreshToken, domain string) {
	formattedDomain := formatDomain(domain)
	ctx.SetCookie("access_token", accessToken, 60*60*24, "/", formattedDomain, true, true)
	ctx.SetCookie("refresh_token", refreshToken, 60*60*24*30, "/", formattedDomain, true, true)
}

// SetRegister sets registration cookies.
func SetRegister(ctx *gin.Context, registerToken, domain string) {
	formattedDomain := formatDomain(domain)
	ctx.SetCookie("register_token", registerToken, 60*60, "/", formattedDomain, true, true)
}

// Clear clears cookies.
func Clear(ctx *gin.Context) {
	ctx.SetCookie("access_token", "", -1, "/", "", true, true)
	ctx.SetCookie("refresh_token", "", -1, "/", "", true, true)
}

// ClearRegister clears registration cookies.
func ClearRegister(ctx *gin.Context) {
	ctx.SetCookie("register_token", "", -1, "/", "", true, true)
}

// ClearAll clears all cookies.
func ClearAll(ctx *gin.Context) {
	Clear(ctx)
	ClearRegister(ctx)
}

// Get gets cookies.
func Get(ctx *gin.Context, key string) (string, error) {
	return ctx.Cookie(key)
}

// GetRegister gets registration cookies.
func GetRegister(ctx *gin.Context, key string) (string, error) {
	return ctx.Cookie(key)
}
