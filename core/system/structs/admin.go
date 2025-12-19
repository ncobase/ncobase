package structs

import "time"

// SystemHealthResponse represents system health status
type SystemHealthResponse struct {
	Status     string                     `json:"status"`
	Timestamp  time.Time                  `json:"timestamp"`
	Version    string                     `json:"version"`
	Uptime     string                     `json:"uptime"`
	Components map[string]ComponentHealth `json:"components"`
}

// ComponentHealth represents health of individual system components
type ComponentHealth struct {
	Status      string            `json:"status"`
	Message     string            `json:"message,omitempty"`
	LastChecked time.Time         `json:"last_checked"`
	Metrics     map[string]string `json:"metrics,omitempty"`
}

// SystemMetricsResponse represents system performance metrics
type SystemMetricsResponse struct {
	TimeRange string          `json:"time_range"`
	CPU       MetricData      `json:"cpu"`
	Memory    MetricData      `json:"memory"`
	Storage   MetricData      `json:"storage"`
	Network   MetricData      `json:"network"`
	Database  DatabaseMetrics `json:"database"`
	API       APIMetrics      `json:"api"`
	Custom    map[string]any  `json:"custom,omitempty"`
}

// MetricData represents time-series metric data
type MetricData struct {
	Current    float64           `json:"current"`
	Average    float64           `json:"average"`
	Peak       float64           `json:"peak"`
	Trend      string            `json:"trend"`
	TimeSeries []TimeSeriesPoint `json:"time_series"`
	Thresholds MetricThresholds  `json:"thresholds"`
}

// TimeSeriesPoint represents a point in time-series data
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// MetricThresholds defines warning and critical thresholds
type MetricThresholds struct {
	Warning  float64 `json:"warning"`
	Critical float64 `json:"critical"`
}

// DatabaseMetrics represents database-specific metrics
type DatabaseMetrics struct {
	Connections     MetricData `json:"connections"`
	QueryTime       MetricData `json:"query_time"`
	SlowQueries     int64      `json:"slow_queries"`
	TransactionRate MetricData `json:"transaction_rate"`
}

// APIMetrics represents API-specific metrics
type APIMetrics struct {
	RequestRate  MetricData       `json:"request_rate"`
	ResponseTime MetricData       `json:"response_time"`
	ErrorRate    MetricData       `json:"error_rate"`
	StatusCodes  map[string]int64 `json:"status_codes"`
	TopEndpoints []EndpointMetric `json:"top_endpoints"`
}

// EndpointMetric represents metrics for individual API endpoints
type EndpointMetric struct {
	Path         string  `json:"path"`
	Method       string  `json:"method"`
	RequestCount int64   `json:"request_count"`
	AvgTime      float64 `json:"avg_time"`
	ErrorRate    float64 `json:"error_rate"`
}

// UserActivityResponse represents paginated user activity logs
type UserActivityResponse struct {
	Activities []ActivityLog `json:"activities"`
	Total      int64         `json:"total"`
	Limit      int           `json:"limit"`
	Offset     int           `json:"offset"`
}

// ActivityLog represents a user activity log entry
type ActivityLog struct {
	ID        string         `json:"id"`
	UserID    string         `json:"user_id"`
	Username  string         `json:"username"`
	Action    string         `json:"action"`
	Resource  string         `json:"resource,omitempty"`
	Details   map[string]any `json:"details,omitempty"`
	IPAddress string         `json:"ip_address"`
	UserAgent string         `json:"user_agent"`
	Timestamp time.Time      `json:"timestamp"`
	Success   bool           `json:"success"`
	ErrorMsg  string         `json:"error_message,omitempty"`
}

// ActivityFilters defines filters for activity log queries
type ActivityFilters struct {
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
	UserID   string `json:"user_id,omitempty"`
	Action   string `json:"action,omitempty"`
	FromDate string `json:"from_date,omitempty"`
	ToDate   string `json:"to_date,omitempty"`
}

// SystemLogsResponse represents paginated system logs
type SystemLogsResponse struct {
	Logs   []SystemLogEntry `json:"logs"`
	Total  int64            `json:"total"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
}

// SystemLogEntry represents a system log entry
type SystemLogEntry struct {
	ID        string         `json:"id"`
	Level     string         `json:"level"`
	Component string         `json:"component"`
	Message   string         `json:"message"`
	Context   map[string]any `json:"context,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

// LogFilters defines filters for system log queries
type LogFilters struct {
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
	Level     string `json:"level,omitempty"`
	Component string `json:"component,omitempty"`
	FromDate  string `json:"from_date,omitempty"`
}

// SystemConfigResponse represents current system configuration
type SystemConfigResponse struct {
	Database     DatabaseConfig    `json:"database"`
	Security     SecurityConfig    `json:"security"`
	Performance  PerformanceConfig `json:"performance"`
	Features     FeatureConfig     `json:"features"`
	Integrations IntegrationConfig `json:"integrations"`
	Maintenance  MaintenanceConfig `json:"maintenance"`
}

// SystemConfigUpdate represents configuration updates
type SystemConfigUpdate struct {
	Security     *SecurityConfigUpdate    `json:"security,omitempty"`
	Performance  *PerformanceConfigUpdate `json:"performance,omitempty"`
	Features     *FeatureConfigUpdate     `json:"features,omitempty"`
	Integrations *IntegrationConfigUpdate `json:"integrations,omitempty"`
	Maintenance  *MaintenanceConfigUpdate `json:"maintenance,omitempty"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	MaxConnections int    `json:"max_connections"`
	IdleTimeout    string `json:"idle_timeout"`
	QueryTimeout   string `json:"query_timeout"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	SessionTimeout    string         `json:"session_timeout"`
	PasswordPolicy    PasswordPolicy `json:"password_policy"`
	TwoFactorEnabled  bool           `json:"two_factor_enabled"`
	LoginAttemptLimit int            `json:"login_attempt_limit"`
	CSRFProtection    bool           `json:"csrf_protection"`
}

// SecurityConfigUpdate represents security configuration updates
type SecurityConfigUpdate struct {
	SessionTimeout    *string         `json:"session_timeout,omitempty"`
	PasswordPolicy    *PasswordPolicy `json:"password_policy,omitempty"`
	TwoFactorEnabled  *bool           `json:"two_factor_enabled,omitempty"`
	LoginAttemptLimit *int            `json:"login_attempt_limit,omitempty"`
	CSRFProtection    *bool           `json:"csrf_protection,omitempty"`
}

// PasswordPolicy represents password policy configuration
type PasswordPolicy struct {
	MinLength        int  `json:"min_length"`
	RequireUppercase bool `json:"require_uppercase"`
	RequireLowercase bool `json:"require_lowercase"`
	RequireNumbers   bool `json:"require_numbers"`
	RequireSymbols   bool `json:"require_symbols"`
}

// PerformanceConfig represents performance configuration
type PerformanceConfig struct {
	CacheEnabled     bool   `json:"cache_enabled"`
	CacheTTL         string `json:"cache_ttl"`
	RateLimitEnabled bool   `json:"rate_limit_enabled"`
	RateLimitRPS     int    `json:"rate_limit_rps"`
}

// PerformanceConfigUpdate represents performance configuration updates
type PerformanceConfigUpdate struct {
	CacheEnabled     *bool   `json:"cache_enabled,omitempty"`
	CacheTTL         *string `json:"cache_ttl,omitempty"`
	RateLimitEnabled *bool   `json:"rate_limit_enabled,omitempty"`
	RateLimitRPS     *int    `json:"rate_limit_rps,omitempty"`
}

// FeatureConfig represents feature toggle configuration
type FeatureConfig struct {
	Analytics         bool `json:"analytics"`
	RealTimeUpdates   bool `json:"real_time_updates"`
	FileSharing       bool `json:"file_sharing"`
	APIProxy          bool `json:"api_proxy"`
	PaymentProcessing bool `json:"payment_processing"`
}

// FeatureConfigUpdate represents feature configuration updates
type FeatureConfigUpdate struct {
	Analytics         *bool `json:"analytics,omitempty"`
	RealTimeUpdates   *bool `json:"real_time_updates,omitempty"`
	FileSharing       *bool `json:"file_sharing,omitempty"`
	APIProxy          *bool `json:"api_proxy,omitempty"`
	PaymentProcessing *bool `json:"payment_processing,omitempty"`
}

// IntegrationConfig represents third-party integration configuration
type IntegrationConfig struct {
	OAuth         OAuthConfig        `json:"oauth"`
	Storage       StorageConfig      `json:"storage"`
	Monitoring    MonitoringConfig   `json:"monitoring"`
	Notifications NotificationConfig `json:"notifications"`
}

// IntegrationConfigUpdate represents integration configuration updates
type IntegrationConfigUpdate struct {
	OAuth         *OAuthConfig        `json:"oauth,omitempty"`
	Storage       *StorageConfig      `json:"storage,omitempty"`
	Monitoring    *MonitoringConfig   `json:"monitoring,omitempty"`
	Notifications *NotificationConfig `json:"notifications,omitempty"`
}

// OAuthConfig represents OAuth configuration
type OAuthConfig struct {
	GoogleEnabled    bool `json:"google_enabled"`
	GitHubEnabled    bool `json:"github_enabled"`
	MicrosoftEnabled bool `json:"microsoft_enabled"`
}

// StorageConfig represents storage configuration
type StorageConfig struct {
	Provider string `json:"provider"`
	Quota    string `json:"quota"`
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	Enabled       bool   `json:"enabled"`
	MetricsLevel  string `json:"metrics_level"`
	RetentionDays int    `json:"retention_days"`
}

// NotificationConfig represents notification configuration
type NotificationConfig struct {
	EmailEnabled   bool `json:"email_enabled"`
	WebhookEnabled bool `json:"webhook_enabled"`
	InAppEnabled   bool `json:"in_app_enabled"`
}

// MaintenanceConfig represents maintenance configuration
type MaintenanceConfig struct {
	MaintenanceMode bool     `json:"maintenance_mode"`
	Message         string   `json:"message,omitempty"`
	AllowedIPs      []string `json:"allowed_ips,omitempty"`
}

// MaintenanceConfigUpdate represents maintenance configuration updates
type MaintenanceConfigUpdate struct {
	MaintenanceMode *bool     `json:"maintenance_mode,omitempty"`
	Message         *string   `json:"message,omitempty"`
	AllowedIPs      *[]string `json:"allowed_ips,omitempty"`
}

// DashboardStatsResponse represents admin dashboard statistics
type DashboardStatsResponse struct {
	Overview       OverviewStats    `json:"overview"`
	Users          UserStats        `json:"users"`
	System         SystemStats      `json:"system"`
	Activity       ActivityStats    `json:"activity"`
	Performance    PerformanceStats `json:"performance"`
	RecentActivity []ActivityLog    `json:"recent_activity"`
	Alerts         []SystemAlert    `json:"alerts"`
}

// OverviewStats represents general overview statistics
type OverviewStats struct {
	TotalUsers    int64   `json:"total_users"`
	ActiveUsers   int64   `json:"active_users"`
	TotalSpaces   int64   `json:"total_spaces"`
	TotalProjects int64   `json:"total_projects"`
	StorageUsed   string  `json:"storage_used"`
	StorageQuota  string  `json:"storage_quota"`
	SystemUptime  string  `json:"system_uptime"`
	HealthScore   float64 `json:"health_score"`
}

// UserStats represents user-related statistics
type UserStats struct {
	NewUsersToday    int64              `json:"new_users_today"`
	NewUsersThisWeek int64              `json:"new_users_this_week"`
	ActiveThisMonth  int64              `json:"active_this_month"`
	UserGrowthTrend  string             `json:"user_growth_trend"`
	TopCountries     []CountryUserCount `json:"top_countries"`
}

// CountryUserCount represents user count by country
type CountryUserCount struct {
	Country string `json:"country"`
	Count   int64  `json:"count"`
}

// SystemStats represents system-related statistics
type SystemStats struct {
	CPUUsage     float64    `json:"cpu_usage"`
	MemoryUsage  float64    `json:"memory_usage"`
	StorageUsage float64    `json:"storage_usage"`
	NetworkIO    string     `json:"network_io"`
	DatabaseSize string     `json:"database_size"`
	BackupStatus string     `json:"backup_status"`
	LastBackup   *time.Time `json:"last_backup,omitempty"`
}

// ActivityStats represents activity-related statistics
type ActivityStats struct {
	RequestsToday       int64    `json:"requests_today"`
	ErrorsToday         int64    `json:"errors_today"`
	AverageResponseTime float64  `json:"average_response_time"`
	TopEndpoints        []string `json:"top_endpoints"`
}

// PerformanceStats represents performance statistics
type PerformanceStats struct {
	ResponseTime    MetricData `json:"response_time"`
	Throughput      MetricData `json:"throughput"`
	ErrorRate       MetricData `json:"error_rate"`
	DatabaseLatency MetricData `json:"database_latency"`
}

// SystemAlert represents system alerts
type SystemAlert struct {
	ID         string         `json:"id"`
	Level      string         `json:"level"`
	Title      string         `json:"title"`
	Message    string         `json:"message"`
	Component  string         `json:"component"`
	Timestamp  time.Time      `json:"timestamp"`
	Resolved   bool           `json:"resolved"`
	ResolvedAt *time.Time     `json:"resolved_at,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// UserManagementResponse represents paginated user management data
type UserManagementResponse struct {
	Users  []UserSummary `json:"users"`
	Total  int64         `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

// UserSummary represents summary user information for management
type UserSummary struct {
	ID           string     `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	DisplayName  string     `json:"display_name"`
	Status       string     `json:"status"`
	Role         string     `json:"role"`
	CreatedAt    time.Time  `json:"created_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	SpaceCount   int        `json:"space_count"`
	ProjectCount int        `json:"project_count"`
	StorageUsed  string     `json:"storage_used"`
}

// UserFilters defines filters for user management queries
type UserFilters struct {
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Search string `json:"search,omitempty"`
	Status string `json:"status,omitempty"`
	Role   string `json:"role,omitempty"`
}

// UserDetailsResponse represents detailed user information
type UserDetailsResponse struct {
	User           UserDetail       `json:"user"`
	Spaces         []SpaceSummary   `json:"spaces"`
	Projects       []ProjectSummary `json:"projects"`
	Sessions       []SessionInfo    `json:"sessions"`
	RecentActivity []ActivityLog    `json:"recent_activity"`
}

// UserDetail represents detailed user information
type UserDetail struct {
	ID               string     `json:"id"`
	Username         string     `json:"username"`
	Email            string     `json:"email"`
	DisplayName      string     `json:"display_name"`
	Avatar           string     `json:"avatar,omitempty"`
	Status           string     `json:"status"`
	Role             string     `json:"role"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	LastLoginAt      *time.Time `json:"last_login_at,omitempty"`
	LoginCount       int64      `json:"login_count"`
	IPAddress        string     `json:"ip_address,omitempty"`
	UserAgent        string     `json:"user_agent,omitempty"`
	EmailVerified    bool       `json:"email_verified"`
	TwoFactorEnabled bool       `json:"two_factor_enabled"`
	StorageUsed      string     `json:"storage_used"`
	StorageQuota     string     `json:"storage_quota"`
}

// SpaceSummary represents space summary information
type SpaceSummary struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	ProjectCount int       `json:"project_count"`
	StorageUsed  string    `json:"storage_used"`
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
}

// ProjectSummary represents project summary information
type ProjectSummary struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	SpaceName    string    `json:"space_name"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
}

// SessionInfo represents user session information
type SessionInfo struct {
	ID        string    `json:"id"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Active    bool      `json:"active"`
}

// UserStatusUpdate represents user status update request
type UserStatusUpdate struct {
	Status string `json:"status" validate:"required,oneof=active inactive suspended"`
	Reason string `json:"reason,omitempty"`
}
