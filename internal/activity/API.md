# Activity Domain API Specification

## Overview
The Activity API provides feed and audit log streams for events across Splittr. It features cursor-based pagination and type-safe discriminated JSON payloads for client apps and AI agents.

---

## Endpoints

### 1. User Activity Feed
- **Method**: `GET /activities`
- **Authentication**: Bearer Auth (Required)
- **Description**: Returns cursor-paginated activities visible to the current user (group activities for groups user belongs to + direct non-group user activity events).

#### Query Parameters
| Parameter | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `limit` | `int` | No | `20` | Items per page (max 100). |
| `cursor` | `string` | No | `""` | Opaque cursor token from previous `pagination.nextCursor`. |

#### Response Schema (200 OK)
```json
{
  "data": [
    {
      "id": "c1f7b84a-9b16-4c92-95f2-491d904b77f1",
      "groupId": "3d5f94b1-8b22-4821-bcf3-54bf93f18e10",
      "actor": {
        "id": "11111111-2222-3333-4444-555555555555",
        "name": "Jane Doe"
      },
      "actionType": "EXPENSE_CREATED",
      "entityType": "EXPENSE",
      "entityId": "a9b8c7d6-e5f4-3210-9876-543210fedcba",
      "description": "added expense 'Dinner' of 45.00 USD",
      "payload": {
        "type": "EXPENSE",
        "expense": { ... },
        "splits": [ ... ]
      },
      "createdAt": "2026-08-01T20:00:00Z"
    }
  ],
  "pagination": {
    "limit": 20,
    "hasMore": false,
    "nextCursor": null
  }
}
```

---

### 2. Group Activity Feed
- **Method**: `GET /groups/{groupId}/feed`
- **Authentication**: Bearer Auth (Required)
- **Description**: Returns cursor-paginated activity log for a specific group. User must be an active group member.

---

## Polymorphic Payload Discriminator (`payload.type`)

The `payload` field is a polymorphic object discriminated by `entityType` and `payload.type`:

### A. `EXPENSE` Payload (`payload.type = "EXPENSE"`)
```json
{
  "type": "EXPENSE",
  "expense": {
    "id": "string",
    "description": "string",
    "amount": 100.0,
    "currency": "USD"
  },
  "splits": [ ... ]
}
```

### B. `SETTLEMENT` Payload (`payload.type = "SETTLEMENT"`)
```json
{
  "type": "SETTLEMENT",
  "expense": {
    "id": "string",
    "amount": 50.0,
    "currency": "USD"
  },
  "split": { ... }
}
```

### C. `MEMBER` Payload (`payload.type = "MEMBER"`)
```json
{
  "type": "MEMBER",
  "member": {
    "userId": "string",
    "role": "admin | member",
    "status": "active"
  }
}
```

### D. `GROUP` Payload (`payload.type = "GROUP"`)
```json
{
  "type": "GROUP",
  "group": {
    "id": "string",
    "name": "string",
    "code": "string"
  },
  "members": [ ... ]
}
```
