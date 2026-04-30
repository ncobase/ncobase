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
