# Feature Interactions

This document records cross-module behavior that must be considered when implementing or changing a
feature.

## Auth, Space, RBAC, and Menus

### Current Flow

1. User logs in through `/login`.
2. Backend builds token payload with user id, global roles, active-space roles, permissions, admin
   state, active space, and available spaces.
3. Frontend stores tokens and fetches `/account`.
4. Frontend requests `/sys/menus/navigation`.
5. Menu tree is filtered by frontend permissions and feature exposure.
6. Requests include `x-md-sid` when a current space is known.

### Current Implemented Behavior

- Space switch must be treated as an authorization change:
  - update local current space,
  - refresh or reissue token for that space domain through `/refresh-token`,
  - preserve only valid user-owned requested spaces and fall back to the default space otherwise,
  - reset permission and request runtime state,
  - refetch `/account`,
  - refetch navigation menus,
  - cancel/remove/invalidate domain queries that depend on ownership or space.
- RBAC changes invalidate menu/account-related caches, reset permission/request runtime state, and
  emit `rbac-change` in the current browser.

### Remaining Behavior

- Menu permissions must be validated against real backend permissions before production seed.
- Affected users in other browsers or devices need a live permission refresh event or must refresh
  their token/re-login.
- RBAC/menu/space membership mutations still need first-class audit display and affected-user/menu
  impact views.

## System Configuration to Runtime Navigation

Menu, dictionary, and option changes affect runtime UI:

- `/sys/menus/navigation` stays available to authenticated users, but raw menu configuration
  (`/sys/menus`, tree, authorized-user lookup, raw get) requires `manage:menu`.
- Menu write operations must invalidate navigation cache.
- Dictionary management reads require `read:dictionaries`, `manage:dictionary`, or
  `manage:system`; writes require `manage:dictionary`.
- Dictionary/option changes can affect forms across system, content, space, payment, resource, AI,
  and builder.
- Option metadata (`category`, `description`, `validation`, `is_secret`, and
  `environment_variable`) is persisted by the backend. The console option category view and list
  filters must use real option rows, with name-prefix fallback only for legacy rows that do not yet
  have a category.
- `is_secret` is a handling flag, not permission to store secret material in `/sys/options`. Secret
  values should remain in environment variables or a secret store; option rows may reference the
  server-side environment variable name.
- Deletes and prefix deletes need usage queries before execution.
- Navigation preview should show hidden, disabled, feature-hidden, and permission-hidden states.

## Config Layering and Runtime Policy

`config.yaml` is limited to bootstrap, infrastructure, secrets, and security-boundary configuration:

- app identity, environment, server/grpc binding, discovery, observability, logger bootstrap,
  extension loading, data stores, queues, search, object storage credentials, email/OAuth secrets,
  JWT secret, Casbin model, and authentication whitelist.

Runtime/product policy must live in system options:

- `system.frontend` controls public frontend URLs used by auth and password reset emails.
- `auth.token` controls access, refresh, register, and MFA token expiry.
- `auth.session` controls maximum sessions, session expiry, and cleanup interval.
- `resource.upload`, `resource.image`, and `resource.quota` control file validation, image
  processing, and quota enforcement.
- `system.storage_policy` controls resource sharing gates and file access audit events; provider
  credentials remain in YAML or a secret store.
- `system.email_policy` controls whether auth and password reset emails are sent; provider
  credentials remain in YAML or a secret store.
- `ai.provider`, `ai.model`, `ai.policy`, `ai.safety`, `ai.cost`, and `ai.embedding` control the AI
  gateway. Provider option values store server-side environment variable names, not secret values.
  Runtime calls resolve secrets on the backend only.

Initialization must ensure missing default options individually and must not skip new defaults just
because older option rows already exist. Existing option values are preserved.

## Content, Resource, and Distribution

### Content Production Chain

```text
taxonomy -> topic -> media/resource -> channel -> distribution -> publish/cancel
```

### Current Reality

- Topics, taxonomies, channels, distributions, media, and topic-media exist in backend.
- Topic create/update persists content metadata directly on `/cms/topics`, including `content_type`,
  `seo_title`, `seo_description`, `seo_keywords`, `excerpt`, tags, and metadata/extras. The console
  SEO dashboard reads these fields as a current metadata snapshot until a dedicated `/cms/seo`
  audit-history module exists.
- Topic list and count filters must stay aligned for search, title, status, taxonomy, content type,
  private, markdown, and space filters; otherwise pagination totals and space-scoped views become
  misleading.
- CMS media can reference resource files through `resource_id`.
- Resource service is available through a content wrapper.
- Frontend media upload now writes to `/res` first, then creates `/cms/media` with the resource
  reference and ownership/space metadata.
- Topic-media synchronization now uses real `/cms/topic-media/by-topic/:topicId` lookup and
  create/update/delete reconciliation.
- Advanced production features such as workflow, template, SEO, version, and schedule are mostly
  frontend-first or absent from current backend modules.

### Required Behavior

- Topic create/edit must validate taxonomy and surface backend errors inline.
- Topic create/edit must expose the real `TopicMediaManager` flow for featured, gallery, and
  attachment media instead of only thumbnail upload.
- Media picker must add existing `/res` selection in addition to upload-created CMS media.
- Resource delete/access/share must check CMS media references now and topic references when the
  reverse topic lookup is added.
- Channel rules should drive distribution creation:
  - `allowed_types` restricts eligible content,
  - `require_review` blocks direct publish,
  - `auto_publish` triggers publish workflow only when allowed.
- Distribution publish/cancel must create audit/activity events and surface failure reasons.

## Resource, Space, and Quota

Resource ownership affects access, quota, and sharing:

- User-owned resource: owner is an individual user.
- Space-owned resource: owner or `space_id` represents a tenant/space.
- Public/shared resource: public routes or token routes bypass normal auth but require rate limit and
  token validation.

Required interactions:

- Upload checks quota before storage write.
- Batch upload reports partial failures and quota stop conditions.
- Space switch changes visible resources and quota scope.
- Resource create/update/delete records or preserves active `space_id` metadata and refreshes both
  owner quota and active-space storage quota from authoritative resource metadata.
- Resource admin cleanup must separate dry-run candidates/potential freed space from actual cleaned
  items/freed space, and must refresh owner/space quota after successful deletes.
- Resource admin stats should aggregate real resource metadata. Download/view activity must be
  marked unavailable until telemetry exists; storage backup must return unavailable until a durable
  backup/export provider is configured.
- Resource admin operations must bypass ownership only through `manage:resources` or
  `admin:resources` and write audit data.

## Payment, Subscription, and Space Billing

Payment and space billing are separate but should converge through events:

1. Product defines commercial item.
2. Channel defines provider and environment config.
3. Order creates payment intent and payment URL.
4. Provider webhook verifies and updates order.
5. Subscription and/or space billing updates after successful payment.
6. Logs and events record every state transition.

Required interactions:

- Payment permissions are now split:
  - `read:payments` for order overview, provider metadata, and stats.
  - `manage:payments` for channels, products, subscriptions, payment URL generation, and verify.
  - `refund:payments` for refunds.
  - `admin:payments` for logs and sensitive administration.
- Webhooks need signature verification, idempotency keys, and retry behavior. Payment log creation and
  response serialization already mask sensitive request/response/error/metadata payload fields.
- Refund must update order state and emit an event.
- Subscription cancel/renew/expire should update space billing if linked.
- Frontend payment detail should show a timeline from order creation through webhook/refund.
- Initialization reruns synchronize the default payment menu and permission seed from legacy
  `read:payment/manage:payment` contracts to the `payments` permission family. Custom Casbin
  policies or manually edited menus still need explicit audit.

## Realtime, Events, Activity, and Notifications

Backend already has event and notification modules. They should become the shared mechanism for:

- resource batch upload progress,
- payment webhook results,
- workflow tasks,
- NCore plugin operation results,
- audit/activity refresh.

Implementation sequence:

1. Header notifications use `/rt/notifications` polling for the current user, including loading,
   error, retry, unread count, and mark-all-read behavior.
2. Backend notification reads and mark-read operations enforce current-user ownership unless the
   caller has realtime management/admin permission.
3. Add WebSocket subscription for high-value live updates.
4. Use event IDs to refresh affected query keys instead of broad polling.
5. Route permissions are split: `read:realtime` for reads/personal notification actions and
   `manage:realtime` or `admin:realtime` for channel management, event publishing/retry/status, and
   system notification writes.

## AI Gateway and Deebus

`deebus` is the provider abstraction library. `ncobase/plugin/ai` is the product gateway and the
trusted boundary for AI calls.

Current behavior:

- AI providers, models, safety, policy, cost, and embedding settings are loaded from `/sys/options`.
- Provider secrets are referenced by environment variable name and resolved only in the backend.
- `/ai/complete`, `/ai/stream`, `/ai/embed`, and `/ai/actions/:action` all create `ncse_ai_run`
  records with user, space, status, provider, model, tokens, duration, cost, error, hashes, and
  sanitized metadata.
- Contextual console assistants for content, resource/media, and Builder call `/ai/actions/:action`
  and return reviewable output. They do not mutate business records directly.
- Normal users can read their own runs and usage. AI managers/admins can inspect cross-user records
  and run provider health checks.

Required interactions:

- Domain modules should call the AI plugin API/service boundary rather than importing `deebus`
  directly.
- Future AI apply flows must define the target domain mutation, permissions, validation, audit, and
  failure behavior before generated output can update records.
- Sensitive metadata and provider errors must remain redacted/truncated according to AI safety
  policy.
- AI usage/cost data should become an operational signal for billing, quota, or alerts only after a
  policy contract is added.

## NCore Operations and Backend Runtime

NCore routes are runtime operations, not normal CRUD:

- They are registered only when management routes are enabled.
- They require `manage:ncore`.
- Load/unload/reload can change available services and plugins.

Required interactions:

- Frontend must probe availability and show local unavailable/forbidden states.
- Mutating operations need confirmation, dependency impact, recent error context, and audit events.
- Backend should record actor, plugin/extension, action, status, duration, error, request id, and
  affected dependencies.

## Builder and Axis

Builder currently generates frontend code and is a developer tool:

- Generated code uses console `createApi` and request behavior.
- Relationship generation distinguishes single-target relations (`oneToOne`, `manyToOne`) from
  collection relations (`oneToMany`, `manyToMany`). Generated table relation columns must call a real
  relation-navigation handler or render inert text, never placeholder console logging.
- Shared components should come from `axis`.
- Backend generation is not complete.

Before production positioning, Builder must generate or document:

- backend schema/migration,
- repository/service/handler,
- route registration,
- permissions and menu seed,
- frontend routes/API/service,
- i18n keys,
- tests and documentation checklist.

## Proxy, Sample, Counter, CLI, Deebus, and Website

- Proxy is a high-risk runtime gateway. `/tbp` management, `/proxy` dynamic routes, and `/ws`
  WebSocket proxying now require authentication plus `manage:tbp`; they still must not be exposed in
  production until endpoint-level ownership/space policy, upstream allowlists, SSRF protection,
  timeout/rate-limit controls, logging redaction, and audit events are in place.
- Sample and counter plugins now require `manage:plugins`; they remain demonstration or internal
  plugin surfaces unless a product owner explicitly promotes them with permissions, menu seed,
  tests, and docs.
- `cli` is an independent scaffolding product. It should only become part of the `ncobase` product
  workflow when Builder or backend generators share a template and verification contract with it.
- `deebus` is now consumed by the `ncobase` AI plugin. New AI behavior must still go through the
  plugin gateway so option keys, secret boundaries, request audit, rate limits, cost tracking, and
  kill switches remain centralized.
- `website` is an independent public site. Shared UI primitives should flow through `axis`, while
  website-specific composition should remain in `website`.
