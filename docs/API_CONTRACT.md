# API Contract Baseline

This is the current contract baseline between `frontend/apps/console` and the `ncobase` backend.
Swagger remains useful, but this file records the implementation reality and known drift.

## Global Contract

| Area | Contract |
| --- | --- |
| Auth header | `Authorization: Bearer <access_token>` |
| Space header | `x-md-sid: <space_id>` |
| JSON fields | `snake_case` |
| Success response | `resp.Success` returns the business payload directly; empty success may be `{ "message": "ok" }`. |
| Error response | `{ code, message, errors }` |
| Upload body | `multipart/form-data`; the browser must set the boundary, so the frontend must not force JSON content type. |
| Permission style | `action:resource`; backend middleware is authoritative. |
| List responses | Not fully uniform yet; each module must document whether it returns `items/total/has_next/next` or another shape. |
| Runtime config | Product/runtime policy lives in `/sys/options`; `config.yaml` is for bootstrap, infrastructure, secrets, and security-boundary config only. |

## Runtime Options

These option names are consumed by backend services at runtime and should be managed through
`/sys/options`, not `config.yaml`.

| Option | Type | Runtime consumers | Notes |
| --- | --- | --- | --- |
| `system.frontend` | object | auth/user email links | `sign_in_url`, `sign_up_url`. |
| `auth.token` | object | auth token generation | `access_token_expiry`, `refresh_token_expiry`, `register_token_expiry`, `mfa_token_expiry`; duration strings such as `2h`, `7d`, `30m`. |
| `auth.session` | object | auth session creation and cleanup | `max_sessions`, `session_expiry`, `cleanup_interval`. |
| `resource.upload` | object | resource upload/batch upload/update/version | `max_upload_size`, `allowed_types`, `default_storage`; storage provider credentials stay in `config.yaml`. |
| `resource.image` | object | resource thumbnails/image processing | thumbnail defaults, max dimensions, compression quality. |
| `resource.quota` | object | resource quota service | `enable_quotas`, `enable_enforcement`, `default_quota`, `warning_threshold`, `quota_check_interval`. |
| `system.storage_policy` | object | resource sharing and access events | `allow_public_links`, `require_owner_scope`, `audit_downloads`; storage provider credentials stay in `config.yaml`. |
| `system.email_policy` | object | auth and password reset email behavior | `enabled`, `allow_auth_email`, `allow_password_reset`, `sender_name`, `digest_frequency`; provider secrets stay in `config.yaml`. |

## Status Labels

- `aligned`: current frontend and backend are expected to work together.
- `partial`: route exists, but permission, response shape, UX, tests, or Swagger still needs work.
- `frontend-only`: frontend wrapper/page exists with no current backend route.
- `backend-only`: backend route exists with no first-class console UI.
- `drift`: implemented route differs from docs, swagger, seed data, or frontend assumptions.

## Auth and Account

| Route | Methods | Frontend caller | Permission | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| `/login` | `POST` | `features/account/apis.ts` | public | aligned | Supports normal login; MFA challenge needs UI completion. |
| `/login/mfa` | `POST` | missing full UI | public | partial | Backend capability exists. |
| `/register` | `POST` | `features/account/apis.ts` | public | aligned | Writes tokens on success. |
| `/logout` | `POST` | `features/account/apis.ts` | authenticated by token if present | aligned | Frontend performs local cleanup even if backend fails. |
| `/refresh-token` | `POST` | `features/account/token_service.ts` | refresh token | aligned | Swagger annotation has been corrected from stale `/refresh`. |
| `/token-status` | `GET` | not first-class | public/current token | backend-only | Useful for diagnostics. |
| `/account` | `GET` | `accountApi.getCurrentUser` | authenticated | aligned | Must return roles, permissions, spaces/default space reliably. |
| `/account/password` | `PUT` | profile TODO | authenticated | partial | Needs password UI and current password policy. |
| `/account/space` | `GET` | account API | authenticated | partial | Space switch write contract still needs confirmation. |
| `/account/spaces` | `GET` | account API and space dropdown | authenticated | aligned | Drives space switcher. |
| `/account/2fa/*` | mixed | missing full UI | authenticated | partial | Setup/verify/disable/backup codes exist. |
| `/sessions` | `GET` | `sessionService` | authenticated | aligned | Session actions need more UI tests. |
| `/sessions/:session_id` | `GET`, `DELETE` | `sessionService` | authenticated | aligned | Current session protection must be verified. |
| `/sessions/deactivate-all` | `POST` | `sessionService` | authenticated | aligned | Requires confirmation in UI. |

## System Domain

| Route family | Methods | Frontend caller | Permission | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| `/sys/users` | CRUD/filter/subroutes | system user APIs | `read/create/update/delete:users`; profile fallbacks | partial | Username/id semantics vary by subroute; add tests. |
| `/sys/employees` | CRUD/list helpers | system user APIs | employee permissions, `manage:hr` | partial | UI coverage must be verified. |
| `/sys/roles` | CRUD, permissions | role/permission APIs | `read:roles`, `manage:roles` | partial | Role changes must trigger permission/token strategy. |
| `/sys/permissions` | CRUD | permission APIs | `super-admin` or `system-admin` role | partial | Frontend guard should reflect backend role gate. |
| `/sys/policies` | CRUD | access APIs | `super-admin` role | partial | Advanced RBAC surface only. |
| `/sys/activities` | create/list/search/get/user | access APIs | authenticated | partial | Audit/activity policy should be tightened. |
| `/sys/menus` | list/get/tree/navigation/authorized + manage ops | menu APIs | reads authenticated, writes `manage:menu` | aligned | Update route is body-style `PUT /sys/menus`. |
| `/sys/dictionaries` | list/get/options/validate/batch + manage ops | dictionary APIs | reads authenticated, writes `manage:dictionary` | aligned | Swagger annotation corrected from stale `/sys/dictionarys`. |
| `/sys/options` | list/get/name/type/batch + manage ops | option APIs | reads authenticated, writes `manage:system` | aligned | Update route is body-style `PUT /sys/options`. |
| `/sys/admin/*` | health/metrics/logs/config/dashboard/users | partial UI/admin pages | `admin:system` | partial | Several services still return placeholder data. |

## Space Domain

| Route family | Methods | Frontend caller | Permission | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| `/sys/spaces` | CRUD | space APIs | group currently requires `manage:spaces` | partial | Read/manage split is needed. |
| `/sys/spaces/:spaceId/users*` | list/add/update/remove/check roles | space user pages | `read:spaces`, `manage:spaces` inside manage group | partial | Top-level group can over-restrict read calls. |
| `/sys/spaces/:spaceId/settings*` | list/get/set/bulk | space APIs | currently under manage group | partial | Public settings route still protected by group. |
| `/sys/spaces/:spaceId/quotas*` | summary/check/update usage | space APIs | currently under manage group | partial | Quota definitions and owner semantics need docs. |
| `/sys/spaces/:spaceId/billing*` | summary/overdue/payment/invoice | space APIs | currently under manage group | partial | Payment plugin relationship not defined. |
| `/sys/spaces/:spaceId/menus|dictionaries|options` | attach/remove/check | space APIs | currently under manage group | partial | Needed for space-level configuration. |

## Content Domain

| Route family | Methods | Frontend caller | Permission | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| `/cms/topics` | CRUD by slug | topic service | authenticated plus Casbin | partial | Frontend routes often use `:id`; settle id/slug. |
| `/cms/taxonomies` | CRUD by slug | taxonomy service | authenticated plus Casbin | partial | Taxonomy relation helpers exist in service layer. |
| `/cms/channels` | CRUD by slug | channel service | authenticated plus Casbin | partial | Channel rules should constrain distribution UI. |
| `/cms/distributions` | CRUD by id, publish/cancel | distribution service | authenticated plus Casbin | partial | Needs state machine and review/schedule integration. |
| `/cms/media` | CRUD by id; list by `resource_id` | media service | authenticated plus Casbin | partial | Media records persist `resource_id` on the media table and list filters also match legacy extras metadata so resource delete checks do not miss older references. |
| `/cms/topic-media` | CRUD and lookup | topic media service | authenticated plus Casbin | backend-only/partial | Frontend topic media logic still contains placeholders. |
| comments/tags/SEO/workflow/templates/versions/schedules | none in current backend | content subfeature APIs | n/a | frontend-only | Hide or implement backend-first. |

## Resource Domain

| Route family | Methods | Frontend caller | Permission | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| `/res` | list/upload | resource APIs | `read:resources` for list, `manage:resources` for upload | aligned | Single upload uses multipart `file`; frontend request layer preserves browser multipart boundaries. |
| `/res/:slug` | get/update/delete | resource APIs | `read:resources`, `manage:resources` | partial | Owner/space/admin rules need integration tests. |
| `/res/:slug/download` | `GET` | resource APIs | `read:resources` | aligned | Blob response. |
| `/res/:slug/versions` | `GET`, `POST` | version history UI | read/manage resources | partial | Version delete/restore semantics need docs. |
| `/res/:slug/access` | `PUT` | resource APIs | `manage:resources` | aligned | Updates `access_level`, `is_public`, and serialized extras together. |
| `/res/:slug/share` | `POST` | share dialog | `manage:resources` | aligned | Accepts `access_level` public/shared plus `expiration_hours`; public returns `/res/dl/:slug`, shared returns tokenized `/res/share/:token`. |
| `/res/view/:slug`, `/res/share/:token`, `/res/thumb/:slug`, `/res/dl/:slug` | `GET` | public URLs | public | partial | Add rate limit and cache policy. |
| `/res/quota`, `/res/usage` | quota and usage summary | resource APIs/admin quota widget | `read:resources` | aligned | Usage returns `usage`, `quota`, `usage_percent`, `quota_exceeded`, `file_count`, and formatted sizes. |
| `/res/batch/*` | batch upload/process/delete/status | resource APIs | read/manage resources | partial | Batch upload uses repeated multipart `files` and supports access level, public flag, path prefix, tags, expiry, processing options, and partial failure result. Batch process/status UI incomplete. |
| `/res/admin/*` | admin files/stats/quotas/jobs/storage | resource admin page | admin | partial | Frontend admin route is admin guarded. |

## Payment Domain

| Route family | Methods | Frontend caller | Permission | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| `/pay/channels` | CRUD/status | payment APIs | Casbin/global, no explicit route middleware | partial | Add `manage:payments` style middleware. |
| `/pay/products` | CRUD | payment APIs | Casbin/global | partial | Product status and billing relation need docs. |
| `/pay/orders` | list/create/get/by-number/payment-url/verify/refund | payment APIs | Casbin/global | partial | Add order state machine and idempotency. |
| `/pay/subscriptions` | list/create/get/update/cancel/by-user | payment APIs | Casbin/global | partial | Add subscription state machine. |
| `/pay/webhooks/:channel` | `POST` | provider callbacks | public/signed | drift/partial | Domain reference corrected; signature/idempotency required. |
| `/pay/logs` | list/get/by-order | payment APIs | Casbin/global | partial | Must mask sensitive payloads. |
| `/pay/providers`, `/pay/stats` | `GET` | payment overview | Casbin/global | partial | Provider utility implementation depth needs review. |

## Realtime and Events

| Route family | Methods | Frontend caller | Permission | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| `/rt/ws` | `GET` WebSocket | not first-class | authenticated | backend-only | Frontend notification center is not wired to WS. |
| `/rt/notifications` | CRUD/read/unread | notification components not fully wired | authenticated | partial | Replace mock notifications. |
| `/rt/channels`, `/rt/events` | channel/event ops | not first-class | authenticated | backend-only | Needed for task progress and audit refresh. |
| `/events`, `/search`, `/stats/realtime` | core event APIs | not first-class | authenticated | backend-only | Route names need documentation and permissions. |

## NCore Management

| Route family | Methods | Frontend caller | Permission | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| `/ncore/*` | extension manager routes | ncore APIs | `manage:ncore` plus auth | aligned/partial | Registered only when backend management routes are enabled; audit events still needed. |

## Known Drift and Required Cleanup

1. Resource menu seed has been aligned to `read:resources/manage:resources`; existing databases may
   need a migration or seed repair.
2. Payment docs previously used `/pay/webhook/:provider`; current route is `/pay/webhooks/:channel`.
3. Content advanced frontend routes do not have current backend modules.
4. Space route groups over-require `manage:spaces` for read operations.
5. Swagger still needs full pass for `/res`, `/tbp`, root auth/account/session, and duplicate route
   warnings.
