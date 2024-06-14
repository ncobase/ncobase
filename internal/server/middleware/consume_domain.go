package middleware

import (
	"stocms/internal/data/structs"
	"stocms/internal/helper"
	"stocms/pkg/consts"
	"stocms/pkg/log"
	"stocms/pkg/resp"

	"github.com/gin-gonic/gin"
)

type DomainFetcher interface {
	UserDomainService(c *gin.Context, username string) (*resp.Exception, error)
}

// ConsumeDomain consumes domain information from the request header or user domains.
func ConsumeDomain(svc DomainFetcher) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve domain ID from request header
		domainID := c.GetHeader(consts.XMdDomainKey)
		// If domain ID is not provided in the header, try to fetch from other sources
		if domainID == "" {
			// Get domain ID
			domainID = helper.GetDomainID(c)
			if domainID == "" {
				// Get user ID
				userID := helper.GetUserID(c)
				// Fetch user domains
				result, _ := svc.UserDomainService(c, userID)
				if result.Code != 0 {
					log.Errorf(nil, "Failed to fetch user domains: %v", result)
				} else if readDomain, ok := result.Data.(*structs.ReadDomain); ok {
					domainID = readDomain.ID
				} else {
					log.Errorf(nil, "Failed to fetch user domains: %v", result)
				}
			}
		}

		// Set domain ID to context
		helper.SetDomainID(c, domainID)

		// Continue to next middleware or handler
		c.Next()
	}
}
