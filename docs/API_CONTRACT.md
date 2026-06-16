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
| `/refresh-token` | `POST` | `features/account/token_service.ts` | refresh token | aligned | Honors a valid requested `x-md-sid` during refresh; returned tokens include active space id, global roles, active space roles, and their permission codes. |
| `/token-status` | `GET` | not first-class | public/current token | backend-only | Useful for diagnostics. |
| `/account` | `GET` | `accountApi.getCurrentUser` | authenticated | aligned | Must return roles, permissions, spaces/default space, and active-space permission context reliably. |
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
| `/sys/roles` | CRUD, permissions | role/permission APIs | `read:roles`, `manage:roles` | partial/aligned | Frontend mutations now trigger local RBAC propagation; backend audit and cross-session live refresh remain future work. |
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
| `/sys/spaces` | CRUD | space APIs | reads `read:spaces`, writes `manage:spaces` | aligned/partial | Group-level over-restriction has been removed; CRUD still needs broader route tests and seed repair notes for existing databases. |
| `/sys/spaces/:spaceId/users*` | list/add/update/remove/check roles | space user pages | reads/checks `read:spaces`, role mutations `manage:spaces` | aligned/partial | Frontend hides add/edit/remove/bulk role actions without `manage:spaces`. |
| `/sys/spaces/:spaceId/settings*` | list/get/set/bulk | space APIs | reads `read:spaces`, set/bulk/create/update/delete `manage:spaces` | partial | Public settings are still behind authenticated space context because they are under `/sys`. |
| `/sys/spaces/:spaceId/quotas*` | summary/check/update usage | space APIs | summary/check/list/get `read:spaces`, usage/config writes `manage:spaces` | partial | Quota definitions and owner semantics need docs. |
| `/sys/spaces/:spaceId/billing*` | summary/overdue/payment/invoice | space APIs | summary/overdue/list/get `read:spaces`, payment/invoice/config writes `manage:spaces` | partial | Payment plugin relationship not defined. |
| `/sys/spaces/:spaceId/menus|dictionaries|options` | attach/remove/check | space APIs | relation reads/checks `read:spaces`, attach/remove `manage:spaces` | partial | Needed for space-level configuration. |

## Content Domain

| Route family | Methods | Frontend caller | Permission | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| `/cms/topics` | CRUD by slug | topic service | authenticated plus Casbin | partial | Frontend routes often use `:id`; settle id/slug. |
| `/cms/taxonomies` | CRUD by slug | taxonomy service | authenticated plus Casbin | partial | Taxonomy relation helpers exist in service layer. |
| `/cms/channels` | CRUD by slug | channel service | authenticated plus Casbin | partial | Channel rules should constrain distribution UI. |
| `/cms/distributions` | CRUD by id, publish/cancel | distribution service | authenticated plus Casbin | partial | Needs state machine and review/schedule integration. |
| `/cms/media` | CRUD by id; list by `resource_id` | media service and content upload hooks | authenticated plus Casbin | partial/aligned | Media records persist `resource_id` plus path, mime, size, owner, and space metadata; list filters also match legacy extras metadata so resource delete checks do not miss older references. Resource enrichment returns preview/download data when available. |
| `/cms/topic-media` | CRUD, list, by-topic, by-topic-and-media | topic media service and `TopicMediaManager` | authenticated plus Casbin | partial/aligned | Frontend calls `/cms/topic-media/by-topic/:topicId`, reconciles create/update/delete differences, and is wired into topic create/edit media fields. Remaining gaps are browser tests and richer reverse reference impact views. |
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
| `/pay/channels` | CRUD/status | payment APIs | `manage:payments` or `admin:payments` | partial/aligned | Channel responses include provider config; do not expose to read-only payment users until masking is implemented. |
| `/pay/products` | CRUD | payment APIs | `manage:payments` or `admin:payments` | partial/aligned | Product status and billing relation need docs. |
| `/pay/orders` | list/get/by-number | payment APIs | `read:payments`, `manage:payments`, `refund:payments`, or `admin:payments` | partial/aligned | Read operations are permission guarded; state machine and idempotency still need tests. |
| `/pay/orders` | create/payment-url/verify | payment APIs | `manage:payments` or `admin:payments` | partial/aligned | Payment URL and verify are operational writes. |
| `/pay/orders/:id/refund` | refund | payment APIs | `refund:payments` or `admin:payments` | partial/aligned | Requires state/amount validation, audit, and provider idempotency tests. |
| `/pay/subscriptions` | list/create/get/update/cancel/by-user | payment APIs | `manage:payments` or `admin:payments` | partial/aligned | Add subscription state machine tests and space billing linkage. |
| `/pay/webhooks/:channel` | `POST` | provider callbacks | public/signed | drift/partial | Domain reference corrected; signature/idempotency required. |
| `/pay/logs` | list/get/by-order | payment APIs | `admin:payments` | aligned/partial | Route is admin-guarded; service masks sensitive request, response, error, and metadata fields on create and response serialization. Export still needs a policy. |
| `/pay/providers`, `/pay/stats` | `GET` | payment overview | `read:payments`, `manage:payments`, `refund:payments`, or `admin:payments` | aligned/partial | Stats aggregate order totals, successful/failed/refunded counts, subscription summary, provider list, selected currency, period bounds, and successful revenue by channel. |

## Proxy, Initialize, Sample, and Counter Plugins

| Route family | Methods | Frontend caller | Permission | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| `/tbp/endpoints`, `/tbp/routes`, `/tbp/transformers` | CRUD | proxy/builder surfaces or future admin UI | global Casbin and menu `manage:tbp` | partial/high-risk | Missing explicit route middleware, ownership/space checks, audit, and transformer safety review. |
| `/proxy/*`, `/ws/*` | dynamic proxy and WebSocket proxy | dynamic clients | configured endpoint/route policy | partial/high-risk | Must validate configured upstream, auth forwarding, rate limit, payload size, timeout, logging redaction, and SSRF protections before production exposure. |
| `/plug/counters` | CRUD | no first-class console product page | authenticated user | backend-only/sample | Built-in counter plugin; add explicit permission or keep as internal/demo surface. |
| `/samples` | CRUD/sample operations | no production console route | none/current handler registration | sample | Demonstration plugin only; should be disabled or guarded in production. |
| `/sys/initialize/status` | initialization status probe | bootstrap/admin operations | status probe | partial/high-risk | Read-only install probe. Do not expose seed execution from this endpoint. |
| `/sys/initialize`, `/sys/initialize/organizations`, `/sys/initialize/users` | seed loading and reruns | bootstrap/admin operations | valid `X-Init-Token` or `manage:system`/`admin:system`/system wildcard | partial/high-risk | Write endpoints accept install token or authenticated system administrator; reruns still need audit and environment guard. |

## Realtime and Events

| Route family | Methods | Frontend caller | Permission | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| `/rt/ws` | `GET` WebSocket | not first-class | `read:realtime`, `manage:realtime`, or `admin:realtime` | backend-only/partial | Header uses polling first; WS subscription still needs frontend wiring. |
| `/rt/notifications` reads and mark read/unread | list/get/read/unread/read-all/unread-all | header notification center | `read:realtime`, `manage:realtime`, or `admin:realtime` | aligned/partial | Read and mark operations are limited to the current user's notifications unless caller has realtime management/admin permission. |
| `/rt/notifications` writes | create/update/delete | no first-class console management page | `manage:realtime` or `admin:realtime` | backend-only/partial | System notification administration needs detail UI, audit, and tests. |
| `/rt/channels` reads and personal subscribe/unsubscribe | list/get/current user channels/subscribe/unsubscribe | not first-class | `read:realtime`, `manage:realtime`, or `admin:realtime` | backend-only/partial | Current user channel route is `/rt/channels/user`. |
| `/rt/channels` management | create/update/delete/subscribers | not first-class | `manage:realtime` or `admin:realtime` | backend-only/partial | Needed for realtime administration and subscriber visibility. |
| `/rt/events` reads | list/history/get | not first-class | `read:realtime`, `manage:realtime`, or `admin:realtime` | backend-only/partial | Static `/history` is registered before `/:id`. |
| `/rt/events` writes | publish/delete | not first-class | `manage:realtime` or `admin:realtime` | backend-only/partial | Event publishing/deletion requires audit and task integration. |
| `/events`, `/search`, `/stats/realtime` | core event APIs | not first-class | read APIs use `read:realtime` or higher; publish/retry/batch/process/status use `manage:realtime` or `admin:realtime` | backend-only/partial | Root event API remains backend-oriented and now has explicit route permissions. |

## NCore Management

| Route family | Methods | Frontend caller | Permission | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| `/ncore/*` | extension manager routes | ncore APIs | `manage:ncore` plus auth | aligned/partial | Registered only when backend management routes are enabled; audit events still needed. |

## Known Drift and Required Cleanup

1. Resource menu seed has been aligned to `read:resources/manage:resources`; existing databases may
   need a migration or seed repair.
2. Payment menu seed has moved from `read:payment/manage:payment` to
   `read:payments/manage:payments/refund:payments/admin:payments`; existing databases need a seed
   repair path.
3. Payment docs previously used `/pay/webhook/:provider`; current route is `/pay/webhooks/:channel`.
4. Content advanced frontend routes do not have current backend modules.
5. CMS media/resource/topic-media are connected, including existing resource selection and topic form
   gallery/attachment UX. Remaining cleanup is browser coverage and richer reverse reference impact
   views.
6. Realtime Swagger still needs a refresh after permission split and `/rt/channels/user` route
   correction.
7. Swagger still needs full pass for `/res`, `/tbp`, root auth/account/session, and duplicate route
   warnings.
