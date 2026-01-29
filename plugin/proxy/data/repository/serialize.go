package repository

import (
	"ncobase/plugin/proxy/data/ent"
	"ncobase/plugin/proxy/structs"
)

// SerializeEndpoint converts ent.Endpoint to structs.ReadEndpoint.
func SerializeEndpoint(row *ent.Endpoint) *structs.ReadEndpoint {
	if row == nil {
		return nil
	}
	return &structs.ReadEndpoint{
		ID:                row.ID,
		Name:              row.Name,
		Description:       row.Description,
		BaseURL:           row.BaseURL,
		Protocol:          row.Protocol,
		AuthType:          row.AuthType,
		AuthConfig:        &row.AuthConfig,
		Timeout:           row.Timeout,
		UseCircuitBreaker: row.UseCircuitBreaker,
		RetryCount:        row.RetryCount,
		ValidateSSL:       row.ValidateSsl,
		LogRequests:       row.LogRequests,
		LogResponses:      row.LogResponses,
		Disabled:          row.Disabled,
		Extras:            &row.Extras,
		CreatedBy:         &row.CreatedBy,
		CreatedAt:         &row.CreatedAt,
		UpdatedBy:         &row.UpdatedBy,
		UpdatedAt:         &row.UpdatedAt,
	}
}

// SerializeEndpoints converts []*ent.Endpoint to []*structs.ReadEndpoint.
func SerializeEndpoints(rows []*ent.Endpoint) []*structs.ReadEndpoint {
	rs := make([]*structs.ReadEndpoint, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, SerializeEndpoint(row))
	}
	return rs
}

// SerializeRoute converts ent.Route to structs.ReadRoute.
func SerializeRoute(row *ent.Route) *structs.ReadRoute {
	if row == nil {
		return nil
	}
	return &structs.ReadRoute{
		ID:                  row.ID,
		Name:                row.Name,
		Description:         row.Description,
		EndpointID:          row.EndpointID,
		PathPattern:         row.PathPattern,
		TargetPath:          row.TargetPath,
		Method:              row.Method,
		InputTransformerID:  &row.InputTransformerID,
		OutputTransformerID: &row.OutputTransformerID,
		CacheEnabled:        row.CacheEnabled,
		CacheTTL:            row.CacheTTL,
		RateLimit:           &row.RateLimit,
		StripAuthHeader:     row.StripAuthHeader,
		Disabled:            row.Disabled,
		Extras:              &row.Extras,
		CreatedBy:           &row.CreatedBy,
		CreatedAt:           &row.CreatedAt,
		UpdatedBy:           &row.UpdatedBy,
		UpdatedAt:           &row.UpdatedAt,
	}
}

// SerializeRoutes converts []*ent.Route to []*structs.ReadRoute.
func SerializeRoutes(rows []*ent.Route) []*structs.ReadRoute {
	rs := make([]*structs.ReadRoute, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, SerializeRoute(row))
	}
	return rs
}

// SerializeTransformer converts ent.Transformer to structs.ReadTransformer.
func SerializeTransformer(row *ent.Transformer) *structs.ReadTransformer {
	if row == nil {
		return nil
	}
	return &structs.ReadTransformer{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		Type:        row.Type,
		Content:     row.Content,
		ContentType: row.ContentType,
		Disabled:    row.Disabled,
		Extras:      &row.Extras,
		CreatedBy:   &row.CreatedBy,
		CreatedAt:   &row.CreatedAt,
		UpdatedBy:   &row.UpdatedBy,
		UpdatedAt:   &row.UpdatedAt,
	}
}

// SerializeTransformers converts []*ent.Transformer to []*structs.ReadTransformer.
func SerializeTransformers(rows []*ent.Transformer) []*structs.ReadTransformer {
	rs := make([]*structs.ReadTransformer, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, SerializeTransformer(row))
	}
	return rs
}

// SerializeLog converts ent.Logs to structs.ReadLog.
func SerializeLog(row *ent.Logs) *structs.ReadLog {
	if row == nil {
		return nil
	}
	return &structs.ReadLog{
		ID:              row.ID,
		EndpointID:      row.EndpointID,
		RouteID:         row.RouteID,
		RequestMethod:   row.RequestMethod,
		RequestPath:     row.RequestPath,
		RequestHeaders:  nil,
		RequestBody:     row.RequestBody,
		StatusCode:      row.StatusCode,
		ResponseHeaders: nil,
		ResponseBody:    row.ResponseBody,
		Duration:        row.Duration,
		Error:           row.Error,
		ClientIP:        row.ClientIP,
		UserID:          row.UserID,
		CreatedAt:       &row.CreatedAt,
	}
}

// SerializeLogs converts []*ent.Logs to []*structs.ReadLog.
func SerializeLogs(rows []*ent.Logs) []*structs.ReadLog {
	rs := make([]*structs.ReadLog, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, SerializeLog(row))
	}
	return rs
}
