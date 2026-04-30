# Domain Reference

API routing conventions and module organization for ncobase.

## Module Structure

### Core Modules (`core/`)

| Module         | Group | Description                                      |
| -------------- | ----- | ------------------------------------------------ |
| `auth`         | root  | Authentication, sessions, MFA, captcha           |
| `access`       | sys   | Role-based access control, permissions, policies |
| `user`         | sys   | User management, profiles, API keys              |
| `organization` | sys   | Organization structure, departments, teams       |
| `space`        | sys   | Multi-tenant spaces, quotas, billing             |
| `system`       | sys   | System configuration, menus, dictionaries        |

### Business Modules (`biz/`)

| Module     | Group | Description                               |
| ---------- | ----- | ----------------------------------------- |
| `content`  | cms   | Articles, topics, media, taxonomies       |
| `realtime` | rt    | Real-time events, notifications, channels |

### Plugins (`plugin/`)

| Plugin     | Group | Description                             |
| ---------- | ----- | --------------------------------------- |
| `resource` | res   | File storage, uploads, quota management |
| `proxy`    | tbp   | API gateway, routing, transformers      |
| `payment`  | pay   | Payment processing, transactions        |
| `counter`  | plug  | Counters, statistics                    |
| `sample`   | plug  | Sample plugin template                  |

## API Routing Patterns

Routes are organized by domain group:

Most business routes use `/{group}/{resource}`. Auth routes are root-level because the auth module
has an empty group.

```text
/{group}/{resource}
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

### Authentication Domain (root)

```text
/login                      # User login
/login/mfa                  # MFA challenge login
/logout                     # User logout
/refresh-token              # Token refresh
/register                   # User registration
/token-status               # Token status check

/account                    # Current account
/account/password           # Current account password
/account/space              # Current default/current space
/account/spaces             # Current account spaces
/account/2fa/status         # MFA status
/account/2fa/setup          # MFA setup
/account/2fa/verify         # MFA verification
/account/2fa/disable        # MFA disable
/account/2fa/backup-codes   # MFA recovery codes

/captcha/generate           # Captcha generation
/captcha/:captcha           # Captcha stream
/captcha/validate           # Captcha validation

/authorize/send             # Send code auth
/authorize/:code            # Complete code auth

/sessions                   # Session list
/sessions/:session_id       # Session detail/delete
/sessions/deactivate-all    # Deactivate other sessions
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

### Realtime Domain (rt)

```
/rt/ws                      # WebSocket endpoint
/rt/notifications           # User notifications
/rt/channels                # Realtime channels
/rt/events                  # Realtime event history
/events                     # Core event publish/detail/search
/stats/realtime             # Realtime stats
```

### Resource Domain (res)

```text
/res                        # File list/upload
/res/:slug                  # File detail/update/delete
/res/:slug/download         # Authenticated download
/res/:slug/versions         # File versions
/res/:slug/share            # Share link generation
/res/:slug/access           # Access level changes
/res/view/:slug             # Public view
/res/share/:token           # Shared file access
/res/thumb/:slug            # Thumbnail access
/res/dl/:slug               # Public download
/res/quota                  # Current user quota
/res/usage                  # Current user usage
/res/batch/upload           # Batch upload
/res/batch/delete           # Batch delete
/res/admin/files            # Admin file management
/res/admin/stats            # Admin storage stats
```

### Proxy Domain (tbp)

```text
/tbp/routes                 # Proxy routes
/tbp/endpoints              # API endpoints
/tbp/transformers           # Request/response transformers
```

### Payment Domain (pay)

```text
/pay/channels               # Payment channels
/pay/orders                 # Payment orders
/pay/orders/:id/payment-url # Payment URL generation
/pay/orders/:id/verify      # Payment verification
/pay/orders/:id/refund      # Refund
/pay/products               # Payment products
/pay/subscriptions          # Subscriptions
/pay/logs                   # Payment logs
/pay/providers              # Provider metadata
/pay/webhooks/:channel      # Channel webhook
/pay/stats                  # Payment stats
```

### NCore Management Domain (ncore)

```text
/ncore/extensions           # Extension inventory
/ncore/plugins/load         # Load plugin
/ncore/plugins/unload       # Unload plugin
/ncore/plugins/reload       # Reload plugin
/ncore/metrics/*            # Runtime metrics
/ncore/health/*             # Runtime health
/ncore/system/*             # Runtime system metadata
```

`/ncore` routes are registered only when extension hot reload is enabled and now require
`manage:ncore` permission in addition to authentication.

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

| Group   | Description         | Examples                  |
| ------- | ------------------- | ------------------------- |
| root    | Authentication      | login, account, sessions  |
| `sys`   | System management   | users, roles, permissions |
| `cms`   | Content management  | topics, media, taxonomies |
| `rt`    | Realtime            | events, notifications     |
| `res`   | Resource management | files, storage, quotas    |
| `pay`   | Payment             | orders, products          |
| `tbp`   | Third-party proxy   | routes, endpoints         |
| `plug`  | Plugin namespace    | custom plugins            |
| `ncore` | Runtime operations  | extensions, health        |
