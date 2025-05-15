package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	spaceStructs "ncobase/space/structs"
	userStructs "ncobase/user/structs"

	"github.com/ncobase/ncore/logging/logger"
)

// RegisterDefaultHooks registers the default processing hooks for the proxy module
func (m *Module) RegisterDefaultHooks() error {
	ctx := context.Background()

	// 1. Example: Create a CRM data synchronization hook set
	// This assumes you have a "salesforce" endpoint and a "contacts" route
	// This would sync contact data with your internal user service
	if err := m.registerCRMSyncHooks(ctx); err != nil {
		logger.Errorf(ctx, "Failed to register CRM sync hooks: %v", err)
		return err
	}

	// 2. Example: Create a payment processing hook set
	// This assumes you have a "stripe" endpoint and a "payments" route
	// This would process payment data with your internal tenant service
	if err := m.registerPaymentProcessingHooks(ctx); err != nil {
		logger.Errorf(ctx, "Failed to register payment processing hooks: %v", err)
		return err
	}

	// 3. Example: Create a collaboration tool hook set
	// This assumes you have a "slack" endpoint and a "messages" route
	// This would process message data with your internal space service
	if err := m.registerCollaborationHooks(ctx); err != nil {
		logger.Errorf(ctx, "Failed to register collaboration hooks: %v", err)
		return err
	}

	return nil
}

// registerCRMSyncHooks registers hooks for CRM data synchronization
func (m *Module) registerCRMSyncHooks(ctx context.Context) error {
	// Find the endpoint by name
	endpoint, err := m.s.Endpoint.GetByName(ctx, "salesforce")
	if err != nil {
		logger.Infof(ctx, "Salesforce endpoint not found, skipping CRM sync hooks")
		return nil // Not an error, just skip
	}

	// Find the route by name
	route, err := m.s.Route.GetByName(ctx, "contacts")
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

		// Enrich with organization data from tenant service
		if orgID, ok := inputData["organization_id"].(string); ok {
			tenant, err := m.tenantService.Tenant.Get(ctx, orgID)
			if err == nil {
				inputData["organization"] = tenant
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
						user, err := m.userService.User.FindUser(ctx, &userStructs.FindUser{Email: email})
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
	err = m.s.Processor.RegisterHook(endpoint.ID, route.ID, preHook, postHook)
	if err != nil {
		return fmt.Errorf("failed to register CRM sync hooks: %w", err)
	}

	logger.Infof(ctx, "Successfully registered CRM sync hooks for endpoint %s and route %s", endpoint.ID, route.ID)
	return nil
}

// registerPaymentProcessingHooks registers hooks for payment processing
func (m *Module) registerPaymentProcessingHooks(ctx context.Context) error {
	// Find the endpoint by name
	endpoint, err := m.s.Endpoint.GetByName(ctx, "stripe")
	if err != nil {
		logger.Infof(ctx, "Stripe endpoint not found, skipping payment processing hooks")
		return nil // Not an error, just skip
	}

	// Find the route by name
	route, err := m.s.Route.GetByName(ctx, "payments")
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

		// Add tenant billing information
		if tenantID, ok := inputData["tenant_id"].(string); ok {
			tenant, err := m.tenantService.Tenant.Get(ctx, tenantID)
			if err == nil && tenant != nil {
				// Add billing information
				inputData["account_name"] = tenant.Name
				// You might get more billing details from tenant extras or a separate service
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
				// Update tenant billing status
				if tenantID, ok := responseData["tenant_id"].(string); ok {
					logger.Infof(ctx, "Payment succeeded for tenant %s", tenantID)
					// Update tenant billing status in your internal service
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
	err = m.s.Processor.RegisterHook(endpoint.ID, route.ID, preHook, postHook)
	if err != nil {
		return fmt.Errorf("failed to register payment processing hooks: %w", err)
	}

	logger.Infof(ctx, "Successfully registered payment processing hooks for endpoint %s and route %s", endpoint.ID, route.ID)
	return nil
}

// registerCollaborationHooks registers hooks for collaboration tool integration
func (m *Module) registerCollaborationHooks(ctx context.Context) error {
	// Find the endpoint by name
	endpoint, err := m.s.Endpoint.GetByName(ctx, "slack")
	if err != nil {
		logger.Infof(ctx, "Slack endpoint not found, skipping collaboration hooks")
		return nil // Not an error, just skip
	}

	// Find the route by name
	route, err := m.s.Route.GetByName(ctx, "messages")
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

		// Add group and user information
		if groupID, ok := inputData["group_id"].(string); ok {
			group, err := m.spaceService.Group.Get(ctx, &spaceStructs.FindGroup{Group: groupID})
			if err == nil && group != nil {
				inputData["group_name"] = group.Name
			}
		}

		if userID, ok := inputData["user_id"].(string); ok {
			user, err := m.userService.User.GetByID(ctx, userID)
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
	err = m.s.Processor.RegisterHook(endpoint.ID, route.ID, preHook, postHook)
	if err != nil {
		return fmt.Errorf("failed to register collaboration hooks: %w", err)
	}

	logger.Infof(ctx, "Successfully registered collaboration hooks for endpoint %s and route %s", endpoint.ID, route.ID)
	return nil
}
