# Permission Rules

Backend permissions are the trusted access boundary. Frontend guards and menu visibility are only UX
helpers and must match backend middleware, seed data, and token payloads.

## Permission Shape

```text
{action}:{resource}
```

Examples:

- `read:users`
- `create:users`
- `update:users`
- `delete:users`
- `manage:roles`
- `admin:system`
- `manage:ncore`

Wildcards supported by middleware and frontend checks:

- `*:*`
- `action:*`
- `*:resource`
- admin/super wildcard patterns handled by backend middleware.

## Identity Inputs

| Input | Source | Purpose |
| --- | --- | --- |
| Access token | `Authorization: Bearer ...` | User identity, roles, permissions, admin state. |
| Space id | `x-md-sid` | Casbin domain and multi-space data boundary. |
| Session | Auth/session middleware | Session validity, device metadata, cleanup. |
| Account response | `/account` | Frontend fallback for roles, permissions, spaces. |

## Backend Enforcement Layers

1. Global middleware restores user and space context.
2. Global Casbin middleware enforces path/method/action/domain policy for non-whitelisted routes.
3. Route middleware applies business semantics such as `HasPermission`, `HasRole`,
   `HasAnyPermission`, `RequireAdmin`.
4. Service/repository code must still enforce ownership and space isolation.

No frontend check can replace any backend layer.

## Current Role Rules

Admin role recognition currently includes:

- `super-admin`
- `system-admin`
- `company-admin`
- `enterprise-admin`
- `admin`

`super-admin` is the strongest role and is used by policy management. `system-admin` can manage
permissions. The legacy `admin` role is still recognized for compatibility.

## Current Route-Level Permissions

| Domain | Route family | Route-level permission |
| --- | --- | --- |
| NCore | `/ncore/*` | `manage:ncore` plus authenticated user |
| Resource | `/res` list/get/search/quota/usage | `read:resources` |
| Resource | `/res` create/update/delete/share/access/version/batch | `manage:resources` |
| Resource admin | `/res/admin/*` | admin |
| Menu writes | `/sys/menus` writes and operations | `manage:menu` |
| Dictionary writes | `/sys/dictionaries` writes | `manage:dictionary` |
| Option writes | `/sys/options` writes | `manage:system` |
| System admin | `/sys/admin/*` | `admin:system` |
| Roles | `/sys/roles` | `read:roles`, `manage:roles` |
| Permissions | `/sys/permissions` | `super-admin` or `system-admin` role |
| Policies | `/sys/policies` | `super-admin` role |
| Users | `/sys/users` | `read/create/update/delete:users` and profile fallbacks |
| Employees | `/sys/employees` | employee permissions and `manage:hr` |
| Organizations | `/sys/orgs` | `read:organizations`, `manage:organizations` |
| Spaces | `/sys/spaces` | currently grouped under `manage:spaces`, with nested read/manage checks |

## Seed and Menu Rules

- Menu `Perms` must use the same string as route middleware or a documented higher-level permission.
- Builder menus use `manage:builder`.
- Resource menus now use `read:resources` and `manage:resources`, matching resource middleware.
- Example menus are development samples; production exposure should be disabled or hidden.
- Menu visibility does not authorize API access.

## Frontend Guard Alignment

| Route | Guard |
| --- | --- |
| `/system/*` | admin |
| `/ncore/*` | admin plus `manage:ncore` |
| `/spaces/*` | super admin UX guard; backend still controls `manage:spaces` |
| `/builder/*` | `manage:builder` and feature exposure |
| `/example/*` | authenticated and feature exposure |
| `/res/*` | `read:resources`; `/res/admin` also admin |
| `/pay/*` | admin UX guard; backend payment permissions need explicit middleware |

## Required Permission Work

1. Split space read and manage routes so list/detail/member read operations can use `read:spaces`
   without inheriting `manage:spaces`.
2. Add explicit payment permissions such as `read:payments`, `manage:payments`, `refund:payments`,
   and `admin:payments`.
3. Add explicit content permissions or document Casbin-only policy for `/cms`.
4. Add route-level audit requirements for all high-risk operations:
   - resource public/share/delete/batch/admin
   - payment refund/webhook/channel config
   - proxy routes and transformers
   - NCore plugin load/unload/reload
   - RBAC/menu/space membership changes
5. Provide a seed repair path for existing databases whenever permission strings change.

## RBAC Change Propagation

After role, permission, policy, menu, or space-role changes:

1. Backend should update persistent RBAC data and optionally write an activity/audit event.
2. Frontend should invalidate menu/navigation and affected user/role queries.
3. Affected users must refresh token or re-login until backend supports live permission refresh.
4. Space switch must refresh token/permissions for the selected space domain.
