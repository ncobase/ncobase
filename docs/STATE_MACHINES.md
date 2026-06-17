# State Machines

These state machines describe intended behavior for current and planned backend-backed workflows.
Existing code may implement only part of a machine; gaps are called out so implementation can follow
one shared model.

## Session

| State | Meaning | Allowed actions | Next states |
| --- | --- | --- | --- |
| `active` | Session can authenticate requests. | refresh access token, delete/deactivate, expire by time. | `active`, `revoked`, `expired` |
| `revoked` | User or admin ended the session. | none except audit/read. | terminal |
| `expired` | Session exceeded expiry. | none except cleanup/read. | terminal |

Rules:

- Deactivate-all must not revoke the current session unless explicitly requested.
- Refresh token must fail for revoked or expired sessions.

## Resource File

| State | Meaning | Allowed actions | Next states |
| --- | --- | --- | --- |
| `uploading` | Upload or batch upload in progress. | cancel if supported. | `active`, `failed` |
| `active` | File is usable by owner/space. | update metadata, create version, share, change access, delete. | `active`, `archived`, `deleted` |
| `archived` | Hidden from normal list but retained. | restore, delete. | `active`, `deleted` |
| `deleted` | Removed from normal access. | purge if retention supports it. | terminal |
| `failed` | Upload/process failed. | retry, delete failed record. | `uploading`, `deleted` |

Access scope is separate from lifecycle:

- `private`
- `space`
- `public`
- `shared_token`

Rules:

- Public/share routes must validate scope or token.
- Delete should check CMS media references through `resource_id`; topic reverse references must be
  included after the resource -> topic lookup is added.
- Batch operations must record per-file status.

## Topic

| State | Meaning | Allowed actions | Next states |
| --- | --- | --- | --- |
| `draft` | Editable content not published. | edit, attach media, start review, schedule, publish if allowed, delete. | `draft`, `reviewing`, `scheduled`, `published`, `deleted` |
| `reviewing` | Waiting for workflow approval. | approve, reject, edit if policy allows. | `draft`, `approved`, `rejected` |
| `approved` | Approved but not published. | publish, schedule, edit. | `published`, `scheduled`, `draft` |
| `scheduled` | Planned for future publish/distribution. | cancel schedule, edit, run publish job. | `draft`, `published`, `failed` |
| `published` | Visible or distributed. | update, create version, unpublish/archive. | `published`, `archived` |
| `archived` | Not active but retained. | restore, delete. | `draft`, `deleted` |
| `deleted` | Trash/deleted state. | restore or purge if trash exists. | terminal or `draft` |

Current backend mainly supports topic CRUD. Workflow, schedule, version, and trash need backend
implementation before this full machine is enforced.

## Distribution

| State | Meaning | Allowed actions | Next states |
| --- | --- | --- | --- |
| `draft` | Distribution record exists but is not published. | publish, schedule, delete. | `published`, `scheduled`, `deleted` |
| `scheduled` | Publish is queued for a future time. | cancel, run job. | `draft`, `published`, `failed` |
| `published` | Content has been distributed to a channel. | cancel/unpublish, retry sync if provider supports it. | `cancelled`, `failed`, `published` |
| `cancelled` | Distribution was cancelled. | republish if policy allows. | `draft`, `published` |
| `failed` | Publish or cancel failed. | retry, cancel. | `published`, `cancelled`, `failed` |
| `deleted` | Record removed. | none. | terminal |

Current backend supports create/update/delete plus publish/cancel. Schedule, failure retry, and review
preconditions are planned.

## Realtime Notification

| State | Meaning | Allowed actions | Next states |
| --- | --- | --- | --- |
| `unread` (`0`) | Notification has not been acknowledged by the recipient. | get, list, mark read. | `read` |
| `read` (`1`) | Recipient has acknowledged the notification. | get, list, mark unread. | `unread` |

Rules:

- Read/list/mark operations are scoped to the authenticated user unless the caller has
  `manage:realtime`, `admin:realtime`, or a wildcard permission.
- `/rt/notifications/read-all` and `/rt/notifications/unread-all` always apply to the current user.
- System notification create/update/delete requires realtime management/admin permission.
- Header notification polling is the current frontend baseline; WebSocket push is a follow-up.

## Payment Order

| State | Meaning | Allowed actions | Next states |
| --- | --- | --- | --- |
| `pending` | Order created; payment not confirmed. | generate payment URL, verify, cancel, expire. | `pending`, `paid`, `failed`, `cancelled`, `expired` |
| `paid` | Provider confirmed payment. | refund full or partial. | `refunded`, `partial_refunded`, `paid` |
| `failed` | Provider or verification failed. | retry payment if allowed. | `pending`, `failed` |
| `cancelled` | User/admin cancelled before payment. | none. | terminal |
| `expired` | Payment window closed. | recreate order if product allows. | terminal |
| `partial_refunded` | Some amount refunded. | refund remaining amount. | `refunded`, `partial_refunded` |
| `refunded` | Fully refunded. | none except audit/read. | terminal |

Rules:

- Webhooks and verify calls must be idempotent.
- Refund requires authorization, current paid state, amount validation, and audit.
- Route permissions are split from lifecycle state: reads use `read:payments`, operational writes use
  `manage:payments`, refunds use `refund:payments`, and log/admin visibility uses `admin:payments`.
- Payment logs are masked by the backend before storage and again during response serialization for
  legacy rows. Secret, token, card, account, signature, authorization, cookie, API key, and credential
  fields must remain redacted in list/detail/export responses.

## Payment Webhook Processing

| State | Meaning | Allowed actions | Next states |
| --- | --- | --- | --- |
| `received` | Provider callback reached `/pay/webhooks/:channel`. | validate signature, parse event, reject. | `validated`, `rejected` |
| `validated` | Signature and channel match. | check idempotency, map event. | `duplicate`, `processing`, `rejected` |
| `processing` | Order/subscription mutation is running. | update domain record, write log, publish event. | `succeeded`, `failed` |
| `duplicate` | Event id or provider reference was already processed. | write idempotent log. | terminal |
| `succeeded` | Domain state was updated and events/logs were written. | none except audit/read. | terminal |
| `failed` | Validation or mutation failed after receipt. | retry if safe, inspect masked log. | `processing`, terminal |
| `rejected` | Signature, channel, payload, or state precondition failed. | none except masked audit/read. | terminal |

Rules:

- The webhook route is public at the auth layer because providers cannot send user tokens.
- Signature validation, channel lookup, idempotency, and payload masking are mandatory before the
  event can mutate orders or subscriptions.
- Retry must be based on provider event id or a generated idempotency key, not on raw request replay.

## AI Run

| State | Meaning | Allowed actions | Next states |
| --- | --- | --- | --- |
| `running` | Completion, stream, embedding, or action call has been accepted and is executing. | observe, stream chunks, provider call completion, provider call failure. | `succeeded`, `failed`, `canceled` |
| `succeeded` | Provider call completed and usage metadata was stored. | read run, aggregate usage, inspect hashes/metadata. | terminal |
| `failed` | Policy, provider, stream, or persistence failure was recorded. | read sanitized error, retry by creating a new run. | terminal |
| `canceled` | Caller or backend canceled execution before completion. | read cancellation metadata. | terminal |

Rules:

- Run modes are `complete`, `stream`, `embed`, and `action`.
- `/ai/actions/:action` stores runs with `mode=action`; `/ai/complete` stores `mode=complete`;
  `/ai/stream` stores `mode=stream`; `/ai/embed` stores `mode=embed`.
- Raw prompts are not stored by default. Request and response hashes can be stored for audit and
  deduplication without retaining sensitive prompt text.
- Run records include `space_id`, `user_id`, `created_by`, and `updated_by`.
- Normal users can read their own runs; `manage:ai` and `admin:ai` can inspect cross-user runs.
- Provider errors must be sanitized and truncated by `ai.safety.max_error_chars`.

## Subscription

| State | Meaning | Allowed actions | Next states |
| --- | --- | --- | --- |
| `trialing` | Trial period active. | activate, cancel, expire. | `active`, `cancelled`, `expired` |
| `active` | Subscription is active. | renew, cancel, mark past due. | `active`, `cancelled`, `past_due` |
| `past_due` | Payment failed or overdue. | retry payment, cancel, expire. | `active`, `cancelled`, `expired` |
| `cancelled` | User/admin cancelled. | reactivate if policy allows. | `active`, `cancelled` |
| `expired` | End date reached without renewal. | renew if product allows. | `active`, `expired` |

Rules:

- Subscription changes should emit events for space billing if linked.
- Cancel/renew actions must show next billing impact in the UI.

## NCore Plugin Operation

| State | Meaning | Allowed actions | Next states |
| --- | --- | --- | --- |
| `idle` | No operation in progress. | load, unload, reload. | `running` |
| `running` | Operation is executing. | observe, cancel if backend supports it. | `succeeded`, `failed` |
| `succeeded` | Operation completed. | view audit/result. | `idle` |
| `failed` | Operation failed. | retry, view error, rollback if supported. | `running`, `idle` |

Rules:

- Every operation must record actor, action, plugin, dependencies, result, duration, and error.
- Frontend must confirm load/unload/reload and display recent errors.

## Workflow Instance

Workflow is not implemented as a backend module in the current `ncobase` content stack, but planned
content approval should follow this machine:

| State | Meaning | Allowed actions | Next states |
| --- | --- | --- | --- |
| `draft` | Workflow not started. | submit. | `pending` |
| `pending` | Waiting for approver. | approve, reject, cancel. | `approved`, `rejected`, `cancelled` |
| `approved` | Approved. | publish/schedule depending on content. | terminal for workflow |
| `rejected` | Rejected with reason. | revise and resubmit. | `draft`, `pending` |
| `cancelled` | Request cancelled. | resubmit if allowed. | `draft`, `pending` |
