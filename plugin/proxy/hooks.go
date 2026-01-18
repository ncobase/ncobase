package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	orgStructs "ncobase/core/organization/structs"
	userStructs "ncobase/core/user/structs"

	"github.com/ncobase/ncore/logging/logger"
)

// RegisterDefaultHooks registers the default processing hooks for the proxy module
func (p *Plugin) RegisterDefaultHooks() error {
	ctx := context.Background()

	// 1. Example: Create a CRM data synchronization hook set
	// This assumes you have a "salesforce" endpoint and a "contacts" route
	// This would sync contact data with your internal user service
	if err := p.registerCRMSyncHooks(ctx); err != nil {
		logger.Errorf(ctx, "Failed to register CRM sync hooks: %v", err)
		return err
	}

	// 2. Example: Create a payment processing hook set
	// This assumes you have a "stripe" endpoint and a "payments" route
	// This would process payment data with your internal space service
	if err := p.registerPaymentProcessingHooks(ctx); err != nil {
		logger.Errorf(ctx, "Failed to register payment processing hooks: %v", err)
		return err
	}

	// 3. Example: Create a collaboration tool hook set
	// This assumes you have a "slack" endpoint and a "messages" route
	// This would process message data with your internal organization service
	if err := p.registerCollaborationHooks(ctx); err != nil {
		logger.Errorf(ctx, "Failed to register collaboration hooks: %v", err)
		return err
	}

	return nil
}

// registerCRMSyncHooks registers hooks for CRM data synchronization
func (p *Plugin) registerCRMSyncHooks(ctx context.Context) error {
	// Find the endpoint by name
	endpoint, err := p.s.Endpoint.GetByName(ctx, "salesforce")
	if err != nil {
		logger.Infof(ctx, "Salesforce endpoint not found, skipping CRM sync hooks")
		return nil // Not an error, just skip
	}

	// Find the route by name
	route, err := p.s.Route.GetByName(ctx, "contacts")
	if err != nil {
		logger.Infof(ctx, "Contacts route not found, skipping CRM sync hooks")
		return nil // Not an error, just skip
	}

	// Create pre-processing hook for CRM data
	preHook := func(ctx context.Context, data []byte) ([]byte, error) {
		// Parse input data
		var inputData map[string]any
		if err := json.Unmarshal(data, &inputData); err != nil {
			return nil, fmt.Errorf("failed to parse input data: %w", err)
		}

		// Enrich with organization data from space service
		if orgID, ok := inputData["org_id"].(string); ok {
			space, err := p.spaceService.Space.Get(ctx, orgID)
			if err == nil {
				inputData["organization"] = space
			}
		}

		// Convert back to JSON
		processedData, err := json.Marshal(inputData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal enriched data: %w", err)
		}

		return processedData, nil
	}

	// Create post-processing hook for CRM data
	postHook := func(ctx context.Context, data []byte) ([]byte, error) {
		// Parse response data
		var responseData map[string]any
		if err := json.Unmarshal(data, &responseData); err != nil {
			return nil, fmt.Errorf("failed to parse response data: %w", err)
		}

		// Process contact data and sync with user service
		if contacts, ok := responseData["contacts"].([]any); ok {
			for _, contact := range contacts {
				if contactMap, ok := contact.(map[string]any); ok {
					// Create or update user from contact data
					// This is simplified, you would need more complex logic in real implementation
					if email, ok := contactMap["email"].(string); ok {
						user, err := p.userService.User.FindUser(ctx, &userStructs.FindUser{Email: email})
						if err == nil {
							// User exists, update with CRM data
							contactMap["user_id"] = user.ID
						} else {
							// User doesn't exist, could create one if needed
							logger.Infof(ctx, "User with email %s not found in system", email)
						}
					}
				}
			}
		}

		// Convert back to JSON
		processedData, err := json.Marshal(responseData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal processed data: %w", err)
		}

		return processedData, nil
	}

	// Register the hooks
	err = p.s.Processor.RegisterHook(endpoint.ID, route.ID, preHook, postHook)
	if err != nil {
		return fmt.Errorf("failed to register CRM sync hooks: %w", err)
	}

	logger.Infof(ctx, "Successfully registered CRM sync hooks for endpoint %s and route %s", endpoint.ID, route.ID)
	return nil
}

// registerPaymentProcessingHooks registers hooks for payment processing
func (p *Plugin) registerPaymentProcessingHooks(ctx context.Context) error {
	// Find the endpoint by name
	endpoint, err := p.s.Endpoint.GetByName(ctx, "stripe")
	if err != nil {
		logger.Infof(ctx, "Stripe endpoint not found, skipping payment processing hooks")
		return nil // Not an error, just skip
	}

	// Find the route by name
	route, err := p.s.Route.GetByName(ctx, "payments")
	if err != nil {
		logger.Infof(ctx, "Payments route not found, skipping payment processing hooks")
		return nil // Not an error, just skip
	}

	// Create pre-processing hook for payment data
	preHook := func(ctx context.Context, data []byte) ([]byte, error) {
		// Parse input data
		var inputData map[string]any
		if err := json.Unmarshal(data, &inputData); err != nil {
			return nil, fmt.Errorf("failed to parse input data: %w", err)
		}

		// Add space billing information
		if spaceID, ok := inputData["space_id"].(string); ok {
			space, err := p.spaceService.Space.Get(ctx, spaceID)
			if err == nil && space != nil {
				// Add billing information
				inputData["account_name"] = space.Name
				// You might get more billing details from space extras or a separate service
			}
		}

		// Convert back to JSON
		processedData, err := json.Marshal(inputData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal enriched payment data: %w", err)
		}

		return processedData, nil
	}

	// Create post-processing hook for payment data
	postHook := func(ctx context.Context, data []byte) ([]byte, error) {
		// Parse response data
		var responseData map[string]any
		if err := json.Unmarshal(data, &responseData); err != nil {
			return nil, fmt.Errorf("failed to parse payment response data: %w", err)
		}

		// Process payment response
		if status, ok := responseData["status"].(string); ok {
			if status == "succeeded" {
				// Update space billing status
				if spaceID, ok := responseData["space_id"].(string); ok {
					logger.Infof(ctx, "Payment succeeded for space %s", spaceID)
					// Update space billing status in your internal service
				}
			} else if status == "failed" {
				// Handle failed payment
				logger.Warnf(ctx, "Payment failed: %v", responseData["error"])
			}
		}

		// Convert back to JSON
		processedData, err := json.Marshal(responseData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal processed payment data: %w", err)
		}

		return processedData, nil
	}

	// Register the hooks
	err = p.s.Processor.RegisterHook(endpoint.ID, route.ID, preHook, postHook)
	if err != nil {
		return fmt.Errorf("failed to register payment processing hooks: %w", err)
	}

	logger.Infof(ctx, "Successfully registered payment processing hooks for endpoint %s and route %s", endpoint.ID, route.ID)
	return nil
}

// registerCollaborationHooks registers hooks for collaboration tool integration
func (p *Plugin) registerCollaborationHooks(ctx context.Context) error {
	// Find the endpoint by name
	endpoint, err := p.s.Endpoint.GetByName(ctx, "slack")
	if err != nil {
		logger.Infof(ctx, "Slack endpoint not found, skipping collaboration hooks")
		return nil // Not an error, just skip
	}

	// Find the route by name
	route, err := p.s.Route.GetByName(ctx, "messages")
	if err != nil {
		logger.Infof(ctx, "Messages route not found, skipping collaboration hooks")
		return nil // Not an error, just skip
	}

	// Create pre-processing hook for message data
	preHook := func(ctx context.Context, data []byte) ([]byte, error) {
		// Parse input data
		var inputData map[string]any
		if err := json.Unmarshal(data, &inputData); err != nil {
			return nil, fmt.Errorf("failed to parse input message data: %w", err)
		}

		// Add organization and user information
		if orgID, ok := inputData["org_id"].(string); ok {
			org, err := p.orgService.Organization.Get(ctx, &orgStructs.FindOrganization{Organization: orgID})
			if err == nil && org != nil {
				inputData["org_name"] = org.Name
			}
		}

		if userID, ok := inputData["user_id"].(string); ok {
			user, err := p.userService.User.GetByID(ctx, userID)
			if err == nil && user != nil {
				inputData["user_name"] = user.Username
				inputData["user_email"] = user.Email
			}
		}

		// Convert back to JSON
		processedData, err := json.Marshal(inputData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal enriched message data: %w", err)
		}

		return processedData, nil
	}

	// Create post-processing hook for message data
	postHook := func(ctx context.Context, data []byte) ([]byte, error) {
		// Parse response data
		var responseData map[string]any
		if err := json.Unmarshal(data, &responseData); err != nil {
			return nil, fmt.Errorf("failed to parse message response data: %w", err)
		}

		// Process message response
		if message, ok := responseData["message"].(map[string]any); ok {
			// Record message in internal system
			logger.Infof(ctx, "Message sent to collaboration tool: %v", message["text"])

			// You might want to store the message ID for future reference
			if messageID, ok := message["id"].(string); ok {
				responseData["internal_message_id"] = messageID
			}
		}

		// Convert back to JSON
		processedData, err := json.Marshal(responseData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal processed message data: %w", err)
		}

		return processedData, nil
	}

	// Register the hooks
	err = p.s.Processor.RegisterHook(endpoint.ID, route.ID, preHook, postHook)
	if err != nil {
		return fmt.Errorf("failed to register collaboration hooks: %w", err)
	}

	logger.Infof(ctx, "Successfully registered collaboration hooks for endpoint %s and route %s", endpoint.ID, route.ID)
	return nil
}
