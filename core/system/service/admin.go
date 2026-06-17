package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	runtimemetrics "runtime/metrics"
	"strconv"
	"strings"
	"time"

	"ncobase/core/system/structs"
	"ncobase/internal/version"

	"github.com/ncobase/ncore/logging/logger"
)

var (
	adminServiceStartedAt          = time.Now()
	ErrAdminAggregationUnavailable = errors.New("admin aggregation is unavailable")
)

// AdminServiceInterface defines admin operations
type AdminServiceInterface interface {
	GetSystemHealth(ctx context.Context) (*structs.SystemHealthResponse, error)
	GetSystemMetrics(ctx context.Context, timeRange string) (*structs.SystemMetricsResponse, error)
	GetUserActivity(ctx context.Context, filters *structs.ActivityFilters) (*structs.UserActivityResponse, error)
	GetSystemLogs(ctx context.Context, filters *structs.LogFilters) (*structs.SystemLogsResponse, error)
	UpdateSystemConfig(ctx context.Context, config *structs.SystemConfigUpdate) (*structs.SystemConfigResponse, error)
	GetSystemConfig(ctx context.Context) (*structs.SystemConfigResponse, error)
	GetDashboardStats(ctx context.Context) (*structs.DashboardStatsResponse, error)
	ManageUsers(ctx context.Context, filters *structs.UserFilters) (*structs.UserManagementResponse, error)
	GetUserDetails(ctx context.Context, userID string) (*structs.UserDetailsResponse, error)
	UpdateUserStatus(ctx context.Context, userID string, statusUpdate *structs.UserStatusUpdate) (map[string]any, error)
}

// adminService implements AdminServiceInterface
type adminService struct {
	s *Service
}

// newAdminService creates admin service
func newAdminService(s *Service) AdminServiceInterface {
	return &adminService{s: s}
}

// GetSystemHealth retrieves comprehensive system health information
func (svc *adminService) GetSystemHealth(ctx context.Context) (*structs.SystemHealthResponse, error) {
	logger.Infof(ctx, "Getting system health information")

	now := time.Now()
	uptime := now.Sub(adminServiceStartedAt)
	components := make(map[string]structs.ComponentHealth)

	dbHealth := structs.ComponentHealth{
		Status:      "healthy",
		Message:     "Database connection active",
		LastChecked: now,
		Metrics:     map[string]string{},
	}
	if err := svc.s.d.Ping(ctx); err != nil {
		dbHealth.Status = "unhealthy"
		dbHealth.Message = fmt.Sprintf("Database connection failed: %v", err)
	} else if db := svc.s.d.GetMasterDB(); db != nil {
		stats := db.Stats()
		dbHealth.Metrics = map[string]string{
			"open_connections": strconv.Itoa(stats.OpenConnections),
			"in_use":           strconv.Itoa(stats.InUse),
			"idle":             strconv.Itoa(stats.Idle),
			"wait_count":       strconv.FormatInt(stats.WaitCount, 10),
			"max_open":         strconv.Itoa(stats.MaxOpenConnections),
		}
	} else {
		dbHealth.Status = "unhealthy"
		dbHealth.Message = "Database connection handle is unavailable"
	}
	components["database"] = dbHealth

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	allocMB := bytesToMB(mem.Alloc)
	sysMB := bytesToMB(mem.Sys)
	memHealth := structs.ComponentHealth{
		Status:      "healthy",
		LastChecked: now,
		Metrics: map[string]string{
			"allocated_mb": fmt.Sprintf("%.2f", allocMB),
			"sys_mb":       fmt.Sprintf("%.2f", sysMB),
			"heap_in_use":  fmt.Sprintf("%.2f MB", bytesToMB(mem.HeapInuse)),
			"gc_runs":      strconv.FormatUint(uint64(mem.NumGC), 10),
		},
	}
	if allocMB > 2048 {
		memHealth.Status = "critical"
		memHealth.Message = "Critical memory usage"
	} else if allocMB > 1024 {
		memHealth.Status = "warning"
		memHealth.Message = "High memory usage detected"
	}
	components["memory"] = memHealth

	cpuHealth := structs.ComponentHealth{
		Status:      "healthy",
		Message:     "Runtime scheduler information available",
		LastChecked: now,
		Metrics: map[string]string{
			"goroutines": strconv.Itoa(runtime.NumGoroutine()),
			"cpu_cores":  strconv.Itoa(runtime.NumCPU()),
		},
	}
	components["cpu"] = cpuHealth

	// Determine overall health status
	overallStatus := "healthy"
	for _, comp := range components {
		if comp.Status == "critical" {
			overallStatus = "critical"
			break
		} else if comp.Status == "warning" && overallStatus != "critical" {
			overallStatus = "warning"
		}
	}

	versionInfo := version.GetVersionInfo()
	return &structs.SystemHealthResponse{
		Status:     overallStatus,
		Timestamp:  now,
		Version:    versionInfo.Version,
		Uptime:     uptime.String(),
		Components: components,
	}, nil
}

// GetSystemMetrics retrieves system performance metrics
func (svc *adminService) GetSystemMetrics(ctx context.Context, timeRange string) (*structs.SystemMetricsResponse, error) {
	logger.Infof(ctx, "Getting system metrics for time range: %s", timeRange)

	var duration time.Duration
	switch timeRange {
	case "1h":
		duration = time.Hour
	case "24h":
		duration = 24 * time.Hour
	case "7d":
		duration = 7 * 24 * time.Hour
	case "30d":
		duration = 30 * 24 * time.Hour
	default:
		duration = 24 * time.Hour
		timeRange = "24h"
	}

	now := time.Now()
	cpuUsage := processCPUPercentSinceStart(now)
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	cpuMetric := structs.MetricData{
		Current:    cpuUsage,
		Average:    cpuUsage,
		Peak:       cpuUsage,
		Trend:      "snapshot",
		TimeSeries: singlePoint(now, cpuUsage),
		Thresholds: structs.MetricThresholds{
			Warning:  70.0,
			Critical: 90.0,
		},
	}

	currentMemoryMB := bytesToMB(memStats.Alloc)
	memoryMetric := structs.MetricData{
		Current:    currentMemoryMB,
		Average:    currentMemoryMB,
		Peak:       currentMemoryMB,
		Trend:      "snapshot",
		TimeSeries: singlePoint(now, currentMemoryMB),
		Thresholds: structs.MetricThresholds{
			Warning:  1024.0,
			Critical: 2048.0,
		},
	}

	dbStats := sql.DBStats{}
	if db := svc.s.d.GetMasterDB(); db != nil {
		dbStats = db.Stats()
	}
	openConnections := float64(dbStats.OpenConnections)
	dbMetrics := structs.DatabaseMetrics{
		Connections: structs.MetricData{
			Current:    openConnections,
			Average:    openConnections,
			Peak:       openConnections,
			Trend:      "snapshot",
			TimeSeries: singlePoint(now, openConnections),
			Thresholds: structs.MetricThresholds{
				Warning:  50,
				Critical: 80,
			},
		},
		QueryTime: structs.MetricData{
			Current: 0,
			Average: 0,
			Peak:    0,
			Trend:   "unavailable",
			Thresholds: structs.MetricThresholds{
				Warning:  100,
				Critical: 500,
			},
		},
		SlowQueries: 0,
		TransactionRate: structs.MetricData{
			Current: 0,
			Average: 0,
			Peak:    0,
			Trend:   "unavailable",
		},
	}

	apiMetrics := structs.APIMetrics{
		RequestRate: structs.MetricData{
			Current: 0,
			Average: 0,
			Peak:    0,
			Trend:   "unavailable",
		},
		ResponseTime: structs.MetricData{
			Current: 0,
			Average: 0,
			Peak:    0,
			Trend:   "unavailable",
			Thresholds: structs.MetricThresholds{
				Warning:  500,
				Critical: 1000,
			},
		},
		ErrorRate: structs.MetricData{
			Current: 0,
			Average: 0,
			Peak:    0,
			Trend:   "unavailable",
			Thresholds: structs.MetricThresholds{
				Warning:  5.0,
				Critical: 10.0,
			},
		},
		StatusCodes:  map[string]int64{},
		TopEndpoints: []structs.EndpointMetric{},
	}

	return &structs.SystemMetricsResponse{
		TimeRange: timeRange,
		CPU:       cpuMetric,
		Memory:    memoryMetric,
		Storage: structs.MetricData{
			Current: 0,
			Average: 0,
			Peak:    0,
			Trend:   "unavailable",
			Thresholds: structs.MetricThresholds{
				Warning:  80.0,
				Critical: 95.0,
			},
		},
		Network: structs.MetricData{
			Current: 0,
			Average: 0,
			Peak:    0,
			Trend:   "unavailable",
		},
		Database: dbMetrics,
		API:      apiMetrics,
		Custom: map[string]any{
			"window_started_at": now.Add(-duration),
			"data_stats":        svc.s.d.GetStats(),
			"note":              "Historical API, storage, and network collectors are not wired.",
		},
	}, nil
}

// GetUserActivity retrieves user activity logs
func (svc *adminService) GetUserActivity(ctx context.Context, filters *structs.ActivityFilters) (*structs.UserActivityResponse, error) {
	logger.Infof(ctx, "Getting user activity with filters: %+v", filters)

	return nil, fmt.Errorf("%w: use /sys/activities for real activity data until admin aggregation is wired", ErrAdminAggregationUnavailable)
}

// GetSystemLogs retrieves system logs
func (svc *adminService) GetSystemLogs(ctx context.Context, filters *structs.LogFilters) (*structs.SystemLogsResponse, error) {
	logger.Infof(ctx, "Getting system logs with filters: %+v", filters)

	return nil, fmt.Errorf("%w: log storage is not configured for /sys/admin/logs", ErrAdminAggregationUnavailable)
}

// UpdateSystemConfig updates system configuration
func (svc *adminService) UpdateSystemConfig(ctx context.Context, configUpdate *structs.SystemConfigUpdate) (*structs.SystemConfigResponse, error) {
	logger.Infof(ctx, "Updating system configuration: %+v", configUpdate)

	return nil, fmt.Errorf("%w: runtime configuration writes must use /sys/options", ErrAdminAggregationUnavailable)
}

// GetSystemConfig retrieves current system configuration
func (svc *adminService) GetSystemConfig(ctx context.Context) (*structs.SystemConfigResponse, error) {
	logger.Infof(ctx, "Getting system configuration")

	securityOption := svc.getObjectOption(ctx, "system.security")
	authTokenOption := svc.getObjectOption(ctx, "auth.token")
	authSessionOption := svc.getObjectOption(ctx, "auth.session")
	performanceOption := svc.getObjectOption(ctx, "system.performance")
	storagePolicy := svc.getObjectOption(ctx, "system.storage_policy")
	emailPolicy := svc.getObjectOption(ctx, "system.email_policy")
	notifications := svc.getObjectOption(ctx, "system.notifications")
	resourceQuota := svc.getObjectOption(ctx, "resource.quota")
	resourceUpload := svc.getObjectOption(ctx, "resource.upload")
	aiProvider := svc.getObjectOption(ctx, "ai.provider")
	aiPolicy := svc.getObjectOption(ctx, "ai.policy")
	maintenance := svc.getObjectOption(ctx, "system.maintenance")
	auditOption := svc.getObjectOption(ctx, "system.audit")
	paymentPolicy := svc.getObjectOption(ctx, "payment.policy")
	proxyPolicy := svc.getObjectOption(ctx, "proxy.policy")

	dbConfig := structs.DatabaseConfig{}
	if db := svc.s.d.GetMasterDB(); db != nil {
		stats := db.Stats()
		dbConfig.MaxConnections = stats.MaxOpenConnections
	}

	return &structs.SystemConfigResponse{
		Database: dbConfig,
		Security: structs.SecurityConfig{
			SessionTimeout: firstString(
				stringFromMap(authSessionOption, "session_expiry"),
				stringFromMap(authTokenOption, "access_token_expiry"),
				minutesFromMap(securityOption, "sessionTimeout"),
			),
			PasswordPolicy: structs.PasswordPolicy{
				MinLength:        intFromMap(securityOption, "passwordMinLength", 0),
				RequireUppercase: boolFromMap(securityOption, "passwordComplexity", false),
				RequireLowercase: boolFromMap(securityOption, "passwordComplexity", false),
				RequireNumbers:   boolFromMap(securityOption, "passwordComplexity", false),
				RequireSymbols:   boolFromMap(securityOption, "requireSymbols", false),
			},
			TwoFactorEnabled:  boolFromMap(securityOption, "mfaRequired", false),
			LoginAttemptLimit: intFromMap(securityOption, "loginAttempts", 0),
			CSRFProtection:    boolFromMap(securityOption, "csrfProtection", false),
		},
		Performance: structs.PerformanceConfig{
			CacheEnabled:     boolFromMap(performanceOption, "cacheEnabled", false),
			CacheTTL:         durationFromSeconds(performanceOption, "cacheTTL"),
			RateLimitEnabled: boolFromNestedMap(performanceOption, "rateLimiting", "enabled", false),
			RateLimitRPS:     intFromNestedMap(performanceOption, "rateLimiting", "requestsPerMinute", 0),
		},
		Features: structs.FeatureConfig{
			Analytics:         boolFromMap(svc.getObjectOption(ctx, "system.audit"), "enabled", false),
			RealTimeUpdates:   boolFromMap(notifications, "in_app", false),
			FileSharing:       boolFromMap(storagePolicy, "allow_public_links", false),
			APIProxy:          boolFromMap(proxyPolicy, "enabled", false),
			PaymentProcessing: boolFromMap(paymentPolicy, "enabled", false),
		},
		Integrations: structs.IntegrationConfig{
			OAuth: structs.OAuthConfig{},
			Storage: structs.StorageConfig{
				Provider: firstString(stringFromMap(storagePolicy, "default_provider"), stringFromMap(resourceUpload, "default_storage")),
				Quota:    bytesStringFromMap(resourceQuota, "default_quota"),
			},
			Monitoring: structs.MonitoringConfig{
				Enabled:       true,
				MetricsLevel:  "runtime",
				RetentionDays: intFromMap(auditOption, "retention", 0),
			},
			Notifications: structs.NotificationConfig{
				EmailEnabled:   boolFromMap(emailPolicy, "enabled", boolFromMap(notifications, "email", false)),
				WebhookEnabled: boolFromMap(notifications, "webhook", false),
				InAppEnabled:   boolFromMap(notifications, "in_app", false),
			},
		},
		Maintenance: structs.MaintenanceConfig{
			MaintenanceMode: boolFromMap(maintenance, "maintenance_mode", false),
			Message:         stringFromMap(maintenance, "message"),
			AllowedIPs:      stringSliceFromMap(maintenance, "allowed_ips"),
		},
		FeaturesMeta: map[string]any{
			"ai_enabled":              boolFromMap(aiProvider, "enabled", false),
			"ai_policy_enabled":       boolFromMap(aiPolicy, "enabled", false),
			"configured_ai_providers": len(sliceFromMap(aiProvider, "providers")),
		},
	}, nil
}

// GetDashboardStats retrieves admin dashboard statistics
func (svc *adminService) GetDashboardStats(ctx context.Context) (*structs.DashboardStatsResponse, error) {
	logger.Infof(ctx, "Getting dashboard statistics")

	return nil, fmt.Errorf("%w: dashboard stats require user, space, resource, and activity aggregators", ErrAdminAggregationUnavailable)
}

// ManageUsers retrieves paginated user list for management
func (svc *adminService) ManageUsers(ctx context.Context, filters *structs.UserFilters) (*structs.UserManagementResponse, error) {
	logger.Infof(ctx, "Getting users for management with filters: %+v", filters)

	return nil, fmt.Errorf("%w: use /sys/users for real user management data until admin aggregation is wired", ErrAdminAggregationUnavailable)
}

// GetUserDetails retrieves detailed user information
func (svc *adminService) GetUserDetails(ctx context.Context, userID string) (*structs.UserDetailsResponse, error) {
	logger.Infof(ctx, "Getting detailed information for user: %s", userID)

	return nil, fmt.Errorf("%w: use /sys/users/%s plus dedicated space/session/activity APIs until admin aggregation is wired", ErrAdminAggregationUnavailable, userID)
}

// UpdateUserStatus updates user status
func (svc *adminService) UpdateUserStatus(ctx context.Context, userID string, statusUpdate *structs.UserStatusUpdate) (map[string]any, error) {
	logger.Infof(ctx, "Updating status for user %s: %+v", userID, statusUpdate)

	return nil, fmt.Errorf("%w: use /sys/users/%s or the dedicated user status API so authorization, audit, and persistence stay in the user module", ErrAdminAggregationUnavailable, userID)
}

func bytesToMB(bytes uint64) float64 {
	return float64(bytes) / 1024 / 1024
}

func singlePoint(ts time.Time, value float64) []structs.TimeSeriesPoint {
	return []structs.TimeSeriesPoint{
		{
			Timestamp: ts,
			Value:     value,
		},
	}
}

func processCPUPercentSinceStart(now time.Time) float64 {
	uptime := now.Sub(adminServiceStartedAt)
	if uptime <= 0 || runtime.NumCPU() <= 0 {
		return 0
	}

	samples := []runtimemetrics.Sample{
		{Name: "/cpu/classes/total:cpu-seconds"},
	}
	runtimemetrics.Read(samples)
	if samples[0].Value.Kind() != runtimemetrics.KindFloat64 {
		return 0
	}

	cpuSeconds := samples[0].Value.Float64()
	if cpuSeconds <= 0 {
		return 0
	}

	usage := (cpuSeconds / uptime.Seconds() / float64(runtime.NumCPU())) * 100
	if usage < 0 {
		return 0
	}
	if usage > 100 {
		return 100
	}
	return usage
}

func (svc *adminService) getObjectOption(ctx context.Context, name string) map[string]any {
	if svc == nil || svc.s == nil || svc.s.Option == nil || strings.TrimSpace(name) == "" {
		return map[string]any{}
	}

	value, err := svc.s.Option.GetObjectByName(ctx, name)
	if err == nil && value != nil {
		return value
	}

	option, optionErr := svc.s.Option.GetByName(ctx, name)
	if optionErr != nil || option == nil || strings.TrimSpace(option.Value) == "" {
		logger.Debugf(ctx, "System option %s is unavailable for admin config snapshot: %v", name, err)
		return map[string]any{}
	}

	var parsed map[string]any
	if jsonErr := json.Unmarshal([]byte(option.Value), &parsed); jsonErr != nil {
		logger.Warnf(ctx, "Failed to parse system option %s as object for admin config snapshot: %v", name, jsonErr)
		return map[string]any{}
	}
	return parsed
}

func firstString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func stringFromMap(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	return stringFromAny(values[key])
}

func stringFromAny(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case json.Number:
		return v.String()
	default:
		return ""
	}
}

func minutesFromMap(values map[string]any, key string) string {
	minutes := intFromMap(values, key, 0)
	if minutes <= 0 {
		return ""
	}
	return fmt.Sprintf("%dm", minutes)
}

func boolFromMap(values map[string]any, key string, defaultValue bool) bool {
	if values == nil {
		return defaultValue
	}
	return boolFromAny(values[key], defaultValue)
}

func boolFromNestedMap(values map[string]any, parent, key string, defaultValue bool) bool {
	nested := mapFromAny(values[parent])
	if nested == nil {
		return defaultValue
	}
	return boolFromMap(nested, key, defaultValue)
}

func boolFromAny(value any, defaultValue bool) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "true", "1", "yes", "y", "on", "enabled":
			return true
		case "false", "0", "no", "n", "off", "disabled":
			return false
		default:
			return defaultValue
		}
	case float64:
		return v != 0
	case float32:
		return v != 0
	case int:
		return v != 0
	case int64:
		return v != 0
	case json.Number:
		parsed, err := strconv.ParseFloat(v.String(), 64)
		if err != nil {
			return defaultValue
		}
		return parsed != 0
	default:
		return defaultValue
	}
}

func intFromMap(values map[string]any, key string, defaultValue int) int {
	if values == nil {
		return defaultValue
	}
	return intFromAny(values[key], defaultValue)
}

func intFromNestedMap(values map[string]any, parent, key string, defaultValue int) int {
	nested := mapFromAny(values[parent])
	if nested == nil {
		return defaultValue
	}
	return intFromMap(nested, key, defaultValue)
}

func intFromAny(value any, defaultValue int) int {
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case int32:
		return int(v)
	case float64:
		return int(v)
	case float32:
		return int(v)
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return defaultValue
		}
		return parsed
	case json.Number:
		parsed, err := strconv.ParseInt(v.String(), 10, 64)
		if err != nil {
			floatParsed, floatErr := strconv.ParseFloat(v.String(), 64)
			if floatErr != nil {
				return defaultValue
			}
			return int(floatParsed)
		}
		return int(parsed)
	default:
		return defaultValue
	}
}

func durationFromSeconds(values map[string]any, key string) string {
	seconds := intFromMap(values, key, 0)
	if seconds <= 0 {
		return ""
	}
	return (time.Duration(seconds) * time.Second).String()
}

func bytesStringFromMap(values map[string]any, key string) string {
	bytes := int64(intFromMap(values, key, 0))
	if bytes <= 0 {
		return ""
	}
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit && exp < 5; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func stringSliceFromMap(values map[string]any, key string) []string {
	items := sliceFromMap(values, key)
	if len(items) == 0 {
		return []string{}
	}

	result := make([]string, 0, len(items))
	for _, item := range items {
		if value := stringFromAny(item); value != "" {
			result = append(result, value)
		}
	}
	return result
}

func sliceFromMap(values map[string]any, key string) []any {
	if values == nil {
		return nil
	}
	switch v := values[key].(type) {
	case []any:
		return v
	case []string:
		result := make([]any, 0, len(v))
		for _, item := range v {
			result = append(result, item)
		}
		return result
	default:
		return nil
	}
}

func mapFromAny(value any) map[string]any {
	switch v := value.(type) {
	case map[string]any:
		return v
	case map[string]string:
		result := make(map[string]any, len(v))
		for key, item := range v {
			result[key] = item
		}
		return result
	default:
		return nil
	}
}
