package service

import (
	"context"
	"ncobase/user/structs"

	"github.com/ncobase/ncore/logging/logger"
)

// UserMeshesServiceInterface defines user meshes service interface
type UserMeshesServiceInterface interface {
	GetUserMeshes(ctx context.Context, username string, includeApiKeys bool) (*structs.UserMeshes, error)
	UpdateUserMeshes(ctx context.Context, username string, updates *structs.UserMeshes) (*structs.UserMeshes, error)
}

// userMeshesService implements UserMeshesServiceInterface
type userMeshesService struct {
	user        UserServiceInterface
	userProfile UserProfileServiceInterface
	employee    EmployeeServiceInterface
	apiKey      ApiKeyServiceInterface
}

// NewUserMeshesService creates new user meshes service
func NewUserMeshesService(
	user UserServiceInterface,
	userProfile UserProfileServiceInterface,
	employee EmployeeServiceInterface,
	apiKey ApiKeyServiceInterface,
) UserMeshesServiceInterface {
	return &userMeshesService{
		user:        user,
		userProfile: userProfile,
		employee:    employee,
		apiKey:      apiKey,
	}
}

// GetUserMeshes retrieves aggregated user information
func (s *userMeshesService) GetUserMeshes(ctx context.Context, username string, includeApiKeys bool) (*structs.UserMeshes, error) {
	// Get basic user info
	user, err := s.user.Get(ctx, username)
	if err != nil {
		return nil, err
	}

	meshes := &structs.UserMeshes{
		User: user,
	}

	// Get user profile
	if profile, err := s.userProfile.Get(ctx, user.ID); err == nil {
		meshes.Profile = profile
	} else {
		logger.Debugf(ctx, "No profile found for user %s: %v", username, err)
	}

	// Get employee info
	if employee, err := s.employee.Get(ctx, user.ID); err == nil {
		meshes.Employee = employee
	} else {
		logger.Debugf(ctx, "No employee record found for user %s: %v", username, err)
	}

	// Get API keys if requested
	if includeApiKeys {
		if apiKeys, err := s.apiKey.GetUserApiKeys(ctx, user.ID); err == nil {
			meshes.ApiKeys = apiKeys
		} else {
			logger.Debugf(ctx, "Could not retrieve API keys for user %s: %v", username, err)
		}
	}

	return meshes, nil
}

// UpdateUserMeshes updates aggregated user information
func (s *userMeshesService) UpdateUserMeshes(ctx context.Context, username string, updates *structs.UserMeshes) (*structs.UserMeshes, error) {
	// Get user to ensure they exist
	user, err := s.user.Get(ctx, username)
	if err != nil {
		return nil, err
	}

	// Update user basic info if provided
	if updates.User != nil {
		userUpdates := map[string]interface{}{}

		if updates.User.Email != "" && updates.User.Email != user.Email {
			userUpdates["email"] = updates.User.Email
		}
		if updates.User.Phone != "" && updates.User.Phone != user.Phone {
			userUpdates["phone"] = updates.User.Phone
		}
		if updates.User.Status != user.Status {
			userUpdates["status"] = updates.User.Status
		}
		if updates.User.IsCertified != user.IsCertified {
			userUpdates["is_certified"] = updates.User.IsCertified
		}
		if updates.User.IsAdmin != user.IsAdmin {
			userUpdates["is_admin"] = updates.User.IsAdmin
		}
		if updates.User.Extras != nil {
			userUpdates["extras"] = updates.User.Extras
		}

		if len(userUpdates) > 0 {
			updatedUser, err := s.user.UpdateUser(ctx, user.ID, userUpdates)
			if err != nil {
				return nil, err
			}
			user = updatedUser
		}
	}

	// Update profile if provided
	var profile *structs.ReadUserProfile
	if updates.Profile != nil {
		profileUpdates := map[string]interface{}{}

		if updates.Profile.DisplayName != "" {
			profileUpdates["display_name"] = updates.Profile.DisplayName
		}
		if updates.Profile.FirstName != "" {
			profileUpdates["first_name"] = updates.Profile.FirstName
		}
		if updates.Profile.LastName != "" {
			profileUpdates["last_name"] = updates.Profile.LastName
		}
		if updates.Profile.Title != "" {
			profileUpdates["title"] = updates.Profile.Title
		}
		if updates.Profile.ShortBio != "" {
			profileUpdates["short_bio"] = updates.Profile.ShortBio
		}
		if updates.Profile.About != nil {
			profileUpdates["about"] = *updates.Profile.About
		}
		if updates.Profile.Thumbnail != nil {
			profileUpdates["thumbnail"] = *updates.Profile.Thumbnail
		}
		if updates.Profile.Links != nil {
			profileUpdates["links"] = *updates.Profile.Links
		}
		if updates.Profile.Extras != nil {
			profileUpdates["extras"] = *updates.Profile.Extras
		}

		if len(profileUpdates) > 0 {
			// Try to update existing profile
			if existingProfile, err := s.userProfile.Get(ctx, user.ID); err == nil {
				updatedProfile, err := s.userProfile.Update(ctx, existingProfile.UserID, profileUpdates)
				if err != nil {
					return nil, err
				}
				profile = updatedProfile
			} else {
				// Create new profile if not exists
				newProfile := &structs.UserProfileBody{
					UserID:      user.ID,
					DisplayName: updates.Profile.DisplayName,
					FirstName:   updates.Profile.FirstName,
					LastName:    updates.Profile.LastName,
					Title:       updates.Profile.Title,
					ShortBio:    updates.Profile.ShortBio,
					About:       updates.Profile.About,
					Thumbnail:   updates.Profile.Thumbnail,
					Links:       updates.Profile.Links,
					Extras:      updates.Profile.Extras,
				}
				createdProfile, err := s.userProfile.Create(ctx, newProfile)
				if err != nil {
					return nil, err
				}
				profile = createdProfile
			}
		}
	}

	// Update employee info if provided
	var employee *structs.ReadEmployee
	if updates.Employee != nil {
		employeeUpdates := &structs.UpdateEmployeeBody{
			EmployeeBody: structs.EmployeeBody{
				UserID:          user.ID,
				TenantID:        updates.Employee.TenantID,
				EmployeeID:      updates.Employee.EmployeeID,
				Department:      updates.Employee.Department,
				Position:        updates.Employee.Position,
				ManagerID:       updates.Employee.ManagerID,
				HireDate:        updates.Employee.HireDate,
				TerminationDate: updates.Employee.TerminationDate,
				EmploymentType:  updates.Employee.EmploymentType,
				Status:          updates.Employee.Status,
				Salary:          updates.Employee.Salary,
				WorkLocation:    updates.Employee.WorkLocation,
				ContactInfo:     updates.Employee.ContactInfo,
				Skills:          updates.Employee.Skills,
				Certifications:  updates.Employee.Certifications,
				Extras:          updates.Employee.Extras,
			},
		}

		// Try to update existing employee record
		if existingEmployee, err := s.employee.Get(ctx, user.ID); err == nil {
			updatedEmployee, err := s.employee.Update(ctx, existingEmployee.UserID, employeeUpdates)
			if err != nil {
				return nil, err
			}
			employee = updatedEmployee
		} else {
			// Create new employee record if not exists
			createEmployeeBody := &structs.CreateEmployeeBody{
				EmployeeBody: employeeUpdates.EmployeeBody,
			}
			createdEmployee, err := s.employee.Create(ctx, createEmployeeBody)
			if err != nil {
				return nil, err
			}
			employee = createdEmployee
		}
	}

	// Return updated meshes
	return &structs.UserMeshes{
		User:     user,
		Profile:  profile,
		Employee: employee,
	}, nil
}
