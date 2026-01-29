package repository

import (
	"ncobase/core/user/data/ent"
	"ncobase/core/user/structs"
)

// SerializeUser converts ent.User to structs.ReadUser.
func SerializeUser(user *ent.User) *structs.ReadUser {
	if user == nil {
		return nil
	}
	return &structs.ReadUser{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Phone:       user.Phone,
		IsCertified: user.IsCertified,
		IsAdmin:     user.IsAdmin,
		Status:      user.Status,
		Extras:      &user.Extras,
		CreatedAt:   &user.CreatedAt,
		UpdatedAt:   &user.UpdatedAt,
	}
}

// SerializeUsers converts ent.User list to structs.ReadUser list.
func SerializeUsers(rows []*ent.User) []*structs.ReadUser {
	result := make([]*structs.ReadUser, 0, len(rows))
	for _, row := range rows {
		result = append(result, SerializeUser(row))
	}
	return result
}

// SerializeEmployee converts ent.Employee to structs.ReadEmployee.
func SerializeEmployee(row *ent.Employee) *structs.ReadEmployee {
	if row == nil {
		return nil
	}
	return &structs.ReadEmployee{
		UserID:          row.ID,
		SpaceID:         row.SpaceID,
		EmployeeID:      row.EmployeeID,
		Department:      row.Department,
		Position:        row.Position,
		ManagerID:       row.ManagerID,
		HireDate:        &row.HireDate,
		TerminationDate: row.TerminationDate,
		EmploymentType:  string(row.EmploymentType),
		Status:          string(row.Status),
		Salary:          row.Salary,
		WorkLocation:    row.WorkLocation,
		ContactInfo:     &row.ContactInfo,
		Skills:          &row.Skills,
		Certifications:  &row.Certifications,
		Extras:          &row.Extras,
		CreatedAt:       &row.CreatedAt,
		UpdatedAt:       &row.UpdatedAt,
	}
}

// SerializeEmployees converts ent.Employee list to structs.ReadEmployee list.
func SerializeEmployees(rows []*ent.Employee) []*structs.ReadEmployee {
	result := make([]*structs.ReadEmployee, 0, len(rows))
	for _, row := range rows {
		result = append(result, SerializeEmployee(row))
	}
	return result
}

// SerializeUserProfile converts ent.UserProfile to structs.ReadUserProfile.
func SerializeUserProfile(row *ent.UserProfile) *structs.ReadUserProfile {
	if row == nil {
		return nil
	}
	return &structs.ReadUserProfile{
		UserID:      row.ID,
		DisplayName: row.DisplayName,
		FirstName:   row.FirstName,
		LastName:    row.LastName,
		Title:       row.Title,
		ShortBio:    row.ShortBio,
		About:       &row.About,
		Thumbnail:   &row.Thumbnail,
		Links:       &row.Links,
		Extras:      &row.Extras,
	}
}

// SerializeApiKey converts ent.ApiKey to structs.ApiKey.
func SerializeApiKey(apiKey *ent.ApiKey) *structs.ApiKey {
	if apiKey == nil {
		return nil
	}
	return &structs.ApiKey{
		ID:        apiKey.ID,
		Name:      apiKey.Name,
		Key:       "",
		UserID:    apiKey.UserID,
		CreatedAt: apiKey.CreatedAt,
		LastUsed:  &apiKey.LastUsed,
	}
}

// SerializeApiKeys converts ent.ApiKey list to structs.ApiKey list.
func SerializeApiKeys(rows []*ent.ApiKey) []*structs.ApiKey {
	result := make([]*structs.ApiKey, 0, len(rows))
	for _, row := range rows {
		result = append(result, SerializeApiKey(row))
	}
	return result
}
