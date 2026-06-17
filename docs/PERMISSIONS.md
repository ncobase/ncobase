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
| Access token | `Authorization: Bearer ...` | User identity, global roles, active-space roles, permissions, and admin state. |
| Space id | `x-md-sid` | Casbin domain, active-space role selection, and multi-space data boundary. |
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
| Resource admin | `/res/admin/*` | `manage:resources` or `admin:resources` |
| Payment | `/pay/orders` list/get, `/pay/providers`, `/pay/stats` | `read:payments`, `manage:payments`, `refund:payments`, or `admin:payments` |
| Payment | `/pay/orders` create/payment-url/verify, `/pay/products`, `/pay/subscriptions`, `/pay/channels` | `manage:payments` or `admin:payments` |
| Payment | `/pay/orders/:id/refund` | `refund:payments` or `admin:payments` |
| Payment | `/pay/logs` | `admin:payments` |
| Payment webhook | `/pay/webhooks/:channel` | public route with provider signature/idempotency requirement |
| AI reads | `/ai/status`, `/ai/providers`, `/ai/models`, `/ai/actions`, `/ai/runs`, `/ai/usage` | `read:ai`, `use:ai`, `manage:ai`, or `admin:ai` |
| AI invocation | `/ai/complete`, `/ai/stream`, `/ai/embed`, `/ai/actions/:action` | `use:ai`, `manage:ai`, or `admin:ai` |
| AI provider health | `/ai/health` | `manage:ai` or `admin:ai` |
| Realtime | `/rt/ws`, `/rt/notifications` reads/mark read, `/rt/channels` reads/personal subscribe, `/rt/events` reads, `/search`, `/stats/realtime` | `read:realtime`, `manage:realtime`, or `admin:realtime` |
| Realtime management | `/rt/notifications` create/update/delete, `/rt/channels` management/subscribers, `/rt/events` publish/delete, `/events` publish/retry/batch/process/status | `manage:realtime` or `admin:realtime` |
| Initialize status | `/sys/initialize/status` | bootstrap status probe |
| Initialize writes | `/sys/initialize*` write endpoints | valid `X-Init-Token` or `manage:system`/`admin:system`/system wildcard |
| Menu navigation | `/sys/menus/navigation` | authenticated user; returned menus still include permission metadata for frontend UX filtering |
| Menu management | `/sys/menus` raw list/get/tree/authorized plus writes and operations | `manage:menu` |
| Dictionary reads | `/sys/dictionaries` list/get/options/validate/batch/usage | `read:dictionaries`, `manage:dictionary`, or `manage:system` |
| Dictionary writes | `/sys/dictionaries` create/update/delete | `manage:dictionary` |
| Option writes | `/sys/options` writes | `manage:system` |
| System admin | `/sys/admin/*` | `admin:system` |
| Roles | `/sys/roles` | `read:roles`, `manage:roles` |
| Permissions | `/sys/permissions` | reads use `read:permissions` or `manage:permissions`; writes use `manage:permissions` |
| Policies | `/sys/policies` | `super-admin` role |
| Activities | `/sys/activities` | reads/search/analytics/types use `read:system` or `manage:system`; create/bulk delete use `manage:system` |
| Users | `/sys/users` | `read/create/update/delete:users`, `manage:users`, and profile fallbacks for owned profile/API-key actions |
| Employees | `/sys/employees` | employee permissions and `manage:hr` |
| Organizations | `/sys/orgs` | `read:organizations`, `manage:organizations` |
| Spaces | `/sys/spaces` | read routes use `read:spaces`; create/update/delete, membership mutations, settings writes, quota writes, billing writes, and relation attach/remove use `manage:spaces` |
| CMS/content | `/cms/*` | reads use `read:cms`, `manage:cms`, `read:content`, or `manage:content`; writes use `manage:cms` or `manage:content` |
| Proxy/TBP | `/tbp`, `/proxy`, `/ws` | `manage:tbp` |
| Demo plugins | counter and sample plugin routes | `manage:plugins` |

## Seed and Menu Rules

- Menu `Perms` must use the same string as route middleware or a documented higher-level permission.
- Builder menus use `manage:builder`.
- Resource menus now use `read:resources` and `manage:resources`, matching resource middleware.
- Payment menus use `read:payments`, `manage:payments`, and `admin:payments`, matching payment route
  middleware. Refund actions use `refund:payments` and are not exposed as a normal navigation menu.
  Every initialization mode that creates the shared payment menus must seed the `payments`
  permission family for the roles that can see those menus.
- Realtime menus use `read:realtime` for notifications/events and `manage:realtime` for channel and
  WebSocket administration. Seed data must provide both permissions for every data mode that exposes
  those menus.
- The primary space navigation menu uses `read:spaces`; explicit management entries such as
  `space-manage` still require `manage:spaces`.
- Example menus are development samples; production exposure should be disabled or hidden.
- Menu visibility does not authorize API access.

## Frontend Guard Alignment

| Route | Guard |
| --- | --- |
| `/system/*` | any matching system submodule permission; subroutes split read and manage actions |
| `/ncore/*` | `manage:ncore` |
| `/spaces/*` | `read:spaces` or `manage:spaces`; create/edit/member mutation subroutes require `manage:spaces` |
| `/builder/*` | `manage:builder` and feature exposure |
| `/example/*` | authenticated and feature exposure |
| `/res/*` | `read:resources`, `manage:resources`, or `admin:resources`; resource admin pages require `manage:resources` or `admin:resources` |
| `/pay/*` | `read:payments`, `manage:payments`, `refund:payments`, or `admin:payments`; product/subscription/channel routes require `manage:payments`; logs require `admin:payments` |
| `/ai/*` | `read:ai`, `use:ai`, `manage:ai`, or `admin:ai`; playground/actions require `use:ai` or higher; provider health requires `manage:ai` or higher |

## Ownership Rules

- `/rt/notifications` list/get/mark-read/mark-unread defaults to the authenticated user. A request
  for another `user_id` is rejected unless the caller has `manage:realtime`, `admin:realtime`, or a
  wildcard permission.
- `/rt/notifications/read-all` and `/rt/notifications/unread-all` always operate on the authenticated
  user, not an arbitrary query/body user.
- `/ai/runs` and `/ai/usage` default to the authenticated user unless the caller has `manage:ai` or
  `admin:ai`. AI run records include `space_id` and `user_id` for audit and tenant scoping.
- `/sys/users/api-keys/:id` delete requires either ownership of the key or elevated user management
  permission (`delete:users`, `manage:users`, admin, or wildcard). `manage:profile` is only enough
  for the current user's own API keys.

## Required Permission Work

1. Provide seed repair for databases that still contain `read:payment` or `manage:payment` menu
   permissions.
2. Provide seed repair for existing databases that still contain enterprise `read:notification`
   instead of `read:realtime`.
3. Provide seed repair for databases that still have legacy `read:organization`,
   `read:user`, `read:permission`, or role-only permission assumptions in menus or custom policies.
4. Add route-level audit requirements for all high-risk operations:
   - resource public/share/delete/batch/admin
   - payment refund/webhook/channel config
   - realtime notification administration, channel management, event publish/retry/status changes
   - proxy routes and transformers
   - NCore plugin load/unload/reload
   - RBAC/menu/space membership changes
5. Provide a seed repair path for existing databases whenever permission strings change.

## RBAC Change Propagation

After role, permission, policy, menu, or space-role changes:

1. Backend should update persistent RBAC data and optionally write an activity/audit event.
2. Frontend mutations reset local permission/request runtime state, invalidate identity and
   navigation queries, invalidate affected user/space records when ids are known, and emit
   `rbac-change`.
3. The current browser receives refreshed account/navigation state after the mutation. Affected users
   in other browsers or devices must refresh token or re-login until backend supports live permission
   refresh events.
4. Space switch must refresh token/permissions for the selected space domain. `/refresh-token`
   honors a valid requested `x-md-sid`, and token payload permissions include global roles plus
   active-space roles.
