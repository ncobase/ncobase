package middleware

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/ctxutil"
	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/security/jwt"
)

var (
	sessionUpdateCache      = sync.Map{}
	sessionUpdateInterval   = 5 * time.Minute
	sessionUpdateCacheMutex = sync.RWMutex{}
)

// SessionMiddleware handles session tracking and updates
func SessionMiddleware(em ext.ManagerInterface) gin.HandlerFunc {

	return func(c *gin.Context) {
		tokenString := extractTokenFromHeader(c)
		if tokenString == "" {
			c.Next()
			return
		}
		// Get Service wrapper manager
		sm := GetServiceManager(em)
		// get access wrapper
		asw := sm.AuthServiceWrapper()
		// Get JWT token manager
		jtm := asw.GetTokenManager()
		if jtm == nil {
			c.Next()
			return
		}

		tokenID := getTokenIDFromJWT(tokenString, jtm)
		if tokenID == "" {
			c.Next()
			return
		}

		// Update session last access time asynchronously
		go updateSessionAccess(asw, tokenID)

		c.Next()
	}
}

// ValidateSessionMiddleware validates if session exists and is active
func ValidateSessionMiddleware(em ext.ManagerInterface) gin.HandlerFunc {

	return func(c *gin.Context) {
		tokenString := extractTokenFromHeader(c)
		if tokenString == "" {
			c.Next()
			return
		}
		// Get Service wrapper manager
		sm := GetServiceManager(em)
		// get access wrapper
		asw := sm.AuthServiceWrapper()
		// Get JWT token manager
		jtm := asw.GetTokenManager()
		if jtm == nil {
			c.Next()
			return
		}

		tokenID := getTokenIDFromJWT(tokenString, jtm)
		if tokenID == "" {
			c.Next()
			return
		}

		// Check session validity
		session, err := asw.GetSessionByTokenID(c.Request.Context(), tokenID)
		if err != nil || !session.IsActive {
			logger.Warnf(c.Request.Context(), "Invalid or inactive session for token: %s", tokenID)
			c.Next()
			return
		}

		// Check expiration
		if session.ExpiresAt != nil && time.Now().UnixMilli() > *session.ExpiresAt {
			go deactivateExpiredSession(asw, tokenID)
			logger.Warnf(c.Request.Context(), "Session expired for token: %s", tokenID)
			c.Next()
			return
		}

		c.Next()
	}
}

// SessionCleanupTask periodically cleans up expired sessions
func SessionCleanupTask(ctx context.Context, em ext.ManagerInterface, interval time.Duration) {
	if interval == 0 {
		interval = 1 * time.Hour
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Get Service wrapper manager
	sm := GetServiceManager(em)
	// get access wrapper
	asw := sm.AuthServiceWrapper()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := asw.CleanupExpiredSessions(ctx); err != nil {
				logger.Errorf(ctx, "Failed to cleanup expired sessions: %v", err)
			}
		}
	}
}

// SessionLimitMiddleware enforces maximum sessions per user
func SessionLimitMiddleware(em ext.ManagerInterface, maxSessions int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := ctxutil.GetUserID(c.Request.Context())
		if userID == "" {
			c.Next()
			return
		}
		// Get Service wrapper manager
		sm := GetServiceManager(em)
		// get access wrapper
		asw := sm.AuthServiceWrapper()
		activeCount := asw.GetActiveSessionsCount(c.Request.Context(), userID)
		if activeCount >= maxSessions {
			logger.Warnf(c.Request.Context(), "User %s exceeded session limit: %d/%d",
				userID, activeCount, maxSessions)
		}

		c.Next()
	}
}

func extractTokenFromHeader(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return ""
	}

	return tokenString
}

func getTokenIDFromJWT(tokenString string, jtm *jwt.TokenManager) string {
	claims, err := jtm.DecodeToken(tokenString)
	if err != nil {
		return ""
	}

	return jwt.GetTokenID(claims)
}

func updateSessionAccess(asw *AuthServiceWrapper, tokenID string) {
	now := time.Now()

	// Check if session was recently updated
	if lastUpdate, ok := sessionUpdateCache.Load(tokenID); ok {
		if lastUpdateTime, ok := lastUpdate.(time.Time); ok {
			if now.Sub(lastUpdateTime) < sessionUpdateInterval {
				return
			}
		}
	}

	// Update database
	ctx := context.Background()
	if err := asw.UpdateSessionLastAccess(ctx, tokenID); err != nil {
		logger.Warnf(ctx, "Failed to update session last access: %v", err)
		return
	}

	// Update cache
	sessionUpdateCache.Store(tokenID, now)
}

func deactivateExpiredSession(asw *AuthServiceWrapper, tokenID string) {
	ctx := context.Background()
	if err := asw.DeactivateSessionByTokenID(ctx, tokenID); err != nil {
		logger.Errorf(ctx, "Failed to deactivate expired session: %v", err)
	}
}
