# Feature Interactions

This document records cross-module behavior that must be considered when implementing or changing a
feature.

## Auth, Space, RBAC, and Menus

### Current Flow

1. User logs in through `/login`.
2. Backend builds token payload with user id, roles, permissions, admin state, default space, and
   available spaces.
3. Frontend stores tokens and fetches `/account`.
4. Frontend requests `/sys/menus/navigation`.
5. Menu tree is filtered by frontend permissions and feature exposure.
6. Requests include `x-md-sid` when a current space is known.

### Required Behavior

- Space switch must be treated as an authorization change:
  - update local current space,
  - refresh or reissue token for that space domain,
  - refetch `/account`,
  - refetch navigation menus,
  - invalidate domain queries that depend on ownership or space.
- RBAC changes must invalidate menu and account-related caches.
- Menu permissions must be validated against real backend permissions before production seed.

## System Configuration to Runtime Navigation

Menu, dictionary, and option changes affect runtime UI:

- Menu write operations must invalidate navigation cache.
- Dictionary/option changes can affect forms across system, content, space, payment, and builder.
- Deletes and prefix deletes need usage queries before execution.
- Navigation preview should show hidden, disabled, feature-hidden, and permission-hidden states.

## Content, Resource, and Distribution

### Content Production Chain

```text
taxonomy -> topic -> media/resource -> channel -> distribution -> publish/cancel
```

### Current Reality

- Topics, taxonomies, channels, distributions, media, and topic-media exist in backend.
- CMS media can reference resource files through `resource_id`.
- Resource service is available through a content wrapper.
- Advanced production features such as workflow, template, SEO, version, and schedule are mostly
  frontend-first or absent from current backend modules.

### Required Behavior

- Topic create/edit must validate taxonomy and surface backend errors inline.
- Topic media picker must connect CMS media to `/res` files.
- Resource delete/access/share must check references once reference query APIs exist.
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
- Resource admin operations must bypass ownership only through admin permission and write audit data.

## Payment, Subscription, and Space Billing

Payment and space billing are separate but should converge through events:

1. Product defines commercial item.
2. Channel defines provider and environment config.
3. Order creates payment intent and payment URL.
4. Provider webhook verifies and updates order.
5. Subscription and/or space billing updates after successful payment.
6. Logs and events record every state transition.

Required interactions:

- Webhooks need signature verification, idempotency keys, retry behavior, and masked logs.
- Refund must update order state and emit an event.
- Subscription cancel/renew/expire should update space billing if linked.
- Frontend payment detail should show a timeline from order creation through webhook/refund.

## Realtime, Events, Activity, and Notifications

Backend already has event and notification modules. They should become the shared mechanism for:

- resource batch upload progress,
- payment webhook results,
- workflow tasks,
- NCore plugin operation results,
- audit/activity refresh.

Implementation sequence:

1. Keep list polling for existing pages.
2. Replace mock notification data with `/rt/notifications`.
3. Add WebSocket subscription for high-value live updates.
4. Use event IDs to refresh affected query keys instead of broad polling.

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
