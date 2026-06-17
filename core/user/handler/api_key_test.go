package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/ctxutil"
)

func TestHasContextPermission(t *testing.T) {
	tests := []struct {
		name        string
		permissions []string
		required    string
		want        bool
	}{
		{name: "exact", permissions: []string{"delete:users"}, required: "delete:users", want: true},
		{name: "action wildcard", permissions: []string{"delete:*"}, required: "delete:users", want: true},
		{name: "resource wildcard", permissions: []string{"*:users"}, required: "delete:users", want: true},
		{name: "global wildcard", permissions: []string{"*:*"}, required: "delete:users", want: true},
		{name: "different permission", permissions: []string{"manage:profile"}, required: "delete:users", want: false},
		{name: "malformed permission", permissions: []string{"delete"}, required: "delete:users", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasContextPermission(tt.permissions, tt.required); got != tt.want {
				t.Fatalf("hasContextPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanManageUserApiKeys(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		permissions []string
		isAdmin     bool
		want        bool
	}{
		{name: "user management", permissions: []string{"manage:users"}, want: true},
		{name: "user delete", permissions: []string{"delete:users"}, want: true},
		{name: "admin context", isAdmin: true, want: true},
		{name: "profile management only", permissions: []string{"manage:profile"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("DELETE", "/sys/users/api-keys/key-1", nil)
			ctx := ctxutil.SetUserPermissions(c.Request.Context(), tt.permissions)
			ctx = ctxutil.SetUserIsAdmin(ctx, tt.isAdmin)
			c.Request = c.Request.WithContext(ctx)

			if got := canManageUserApiKeys(c); got != tt.want {
				t.Fatalf("canManageUserApiKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}
