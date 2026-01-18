# Domain Reference

API routing conventions and module organization for ncobase.

## Module Structure

### Core Modules (`core/`)

| Module         | Group | Description                                      |
|----------------|-------|--------------------------------------------------|
| `access`       | sys   | Role-based access control, permissions, policies |
| `auth`         | sys   | Authentication, sessions, MFA, captcha           |
| `user`         | sys   | User management, profiles, API keys              |
| `organization` | sys   | Organization structure, departments, teams       |
| `space`        | sys   | Multi-tenant spaces, quotas, billing             |
| `system`       | sys   | System configuration, menus, dictionaries        |

### Business Modules (`biz/`)

| Module     | Group | Description                               |
|------------|-------|-------------------------------------------|
| `content`  | cms   | Articles, topics, media, taxonomies       |
| `realtime` | msg   | Real-time events, notifications, channels |

### Plugins (`plugin/`)

| Plugin     | Group | Description                             |
|------------|-------|-----------------------------------------|
| `resource` | res   | File storage, uploads, quota management |
| `proxy`    | api   | API gateway, routing, transformers      |
| `payment`  | pay   | Payment processing, transactions        |
| `counter`  | util  | Counters, statistics                    |
| `sample`   | plug  | Sample plugin template                  |

## API Routing Patterns

Routes are organized by domain group:

```text
/api/{group}/{resource}
```

### System Domain (sys)

```text
/sys/roles                  # Role management
/sys/roles/:slug/permissions
/sys/permissions            # Permission management
/sys/policies               # RBAC policies

/sys/users                  # User management
/sys/users/:id/profile
/sys/users/:id/api-keys
/sys/employees              # Employee records

/sys/orgs                   # Organizations
/sys/orgs/:id/roles
/sys/orgs/:id/members

/sys/spaces                 # Multi-tenant spaces
/sys/spaces/:id/settings
/sys/spaces/:id/quotas
/sys/spaces/:id/billing

/sys/menus                  # System menus
/sys/options                # System options
/sys/dictionaries           # Data dictionaries

/sys/activities             # Activity logs
/sys/activities/search
```

### Authentication Domain (auth)

```text
/auth/login                 # User login
/auth/logout                # User logout
/auth/refresh               # Token refresh
/auth/register              # User registration

/auth/mfa/setup             # MFA setup
/auth/mfa/verify            # MFA verification
/auth/mfa/recovery          # Recovery codes

/auth/captcha               # Captcha generation
/auth/sessions              # Session management
```

### Content Domain (cms)

```text
/cms/topics                 # Topic/article management
/cms/topics/:id/media
/cms/channels               # Content channels
/cms/taxonomies             # Categories/tags
/cms/media                  # Media files
/cms/distributions          # Content distribution
```

### Messaging Domain (msg)

```
/msg/events                 # Real-time events
/msg/channels               # Message channels
/msg/subscriptions          # Event subscriptions
/msg/notifications          # User notifications
```

### Resource Domain (res)

```text
/res/files                  # File management
/res/files/upload
/res/files/:id/download
/res/quotas                 # Storage quotas
```

### API Gateway Domain (api)

```text
/api/routes                 # Proxy routes
/api/endpoints              # API endpoints
/api/transformers           # Request/response transformers
```

## Permission Patterns

```text
{action}:{resource}

read:roles                  # View roles
manage:roles                # Create/update/delete roles
read:users                  # View users
manage:users                # Manage users
admin:system                # System administration
```

## Domain Groups

| Group  | Description         | Examples                  |
|--------|---------------------|---------------------------|
| `sys`  | System management   | users, roles, permissions |
| `auth` | Authentication      | login, mfa, sessions      |
| `cms`  | Content management  | topics, media, taxonomies |
| `msg`  | Messaging/realtime  | events, notifications     |
| `res`  | Resource management | files, storage, quotas    |
| `api`  | API gateway         | routes, endpoints         |
| `plug` | Plugin namespace    | custom plugins            |
